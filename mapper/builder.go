package mapper

import (
	"go/types"
	"sort"
	"strings"

	"github.com/dave/jennifer/jen"
)

// builderTarget is a method that fills a builder instead of returning a struct: its
// output type is also one of its inputs, and the builder's setters are the fields.
type builderTarget struct {
	param  *GoParameter
	fields []*builderField
}

// builderField is one `SetX(v T)` setter, together with the `SetNillableX(v *T)`
// twin ent generates for an optional column when there is one.
type builderField struct {
	value          *GoField
	setter         string
	nillable       *GoField
	nillableSetter string
}

// detectBuilder recognises a builder method: an output type that is also one of
// the inputs and declares at least one setter, a method `SetX(v T)` returning the
// builder itself so that calls chain.
func detectBuilder(f *GoFunction) *builderTarget {
	if len(f.inputs) < 2 {
		return nil
	}

	for _, input := range f.inputs {
		if input.typ.String() != f.output.String() {
			continue
		}

		fields := setterFields(f.output, f.options)
		if len(fields) == 0 {
			return nil
		}

		return &builderTarget{param: input, fields: fields}
	}

	return nil
}

// setterFields turns the setters of builder into target fields, sorted by name so
// the generated method is stable. Two ent conventions are folded in: an optional
// column's `SetNillableX(*T)` is attached to `SetX(T)` rather than being a field of
// its own, and an edge's `SetX(*Entity)` steps aside for `SetXID(id)`, unless a
// directive addresses X and not XID.
func setterFields(builder *GoType, options map[string]*FieldOptions) []*builderField {
	named := namedType(builder)
	if named == nil {
		return nil
	}

	setters := map[string]*GoField{}
	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		name, ok := strings.CutPrefix(method.Name(), "Set")
		if !ok || name == "" || !method.Exported() {
			continue
		}

		sig := method.Type().(*types.Signature)
		if sig.Params().Len() != 1 || sig.Variadic() || sig.Results().Len() != 1 {
			continue
		}
		if types.TypeString(sig.Results().At(0).Type(), nil) != builder.String() {
			continue
		}

		setters[name] = &GoField{name: name, typ: getGoType(sig.Params().At(0).Type())}
	}

	dropped := map[string]bool{}
	for name, field := range setters {
		if base, ok := strings.CutPrefix(name, "Nillable"); ok && setters[base] != nil {
			dropped[name] = true
			continue
		}

		withID := setters[name+"ID"]
		if withID == nil || field.typ.Kind != PointerKind || field.typ.Element.Kind != StructKind {
			continue
		}

		if options[name] != nil && options[name+"ID"] == nil {
			dropped[name+"ID"] = true
		} else {
			dropped[name] = true
		}
	}

	var fields []*builderField
	for name, field := range setters {
		if dropped[name] {
			continue
		}

		bf := &builderField{value: field, setter: "Set" + name}
		if nillable := setters["Nillable"+name]; nillable != nil {
			bf.nillable = nillable
			bf.nillableSetter = "SetNillable" + name
		}

		fields = append(fields, bf)
	}

	sort.Slice(fields, func(i, j int) bool {
		return fields[i].value.name < fields[j].value.name
	})

	return fields
}

// generateBuilderMethod emits a method that reads the other inputs and calls one
// setter per field on the builder, then returns the builder. A nil source leaves
// the builder untouched.
func (m *Generator) generateBuilderMethod(mapperName string, f *GoFunction, target *builderTarget) error {
	// the builder is the sink, every other input is a source
	src := *f
	src.inputs = nil
	for _, input := range f.inputs {
		if input != target.param {
			src.inputs = append(src.inputs, input)
		}
	}

	var resolved []*ResolvedMapping
	setters := map[*ResolvedMapping]string{}
	for _, bf := range target.fields {
		field, setter := bf.value, bf.setter

		if f.skipOptional && bf.nillable != nil && !m.hasSource(&src, field) {
			continue
		}

		// a pointer source keeps its nil through the nillable setter rather than
		// being flattened to a zero value, unless a default: stands in for nil; and
		// a zero:"nil" field hands the nillable setter its nil
		options := f.options[field.name]
		hasFallback := options != nil && options.fallback != ""
		zeroNil := options != nil && options.zeroNil
		if bf.nillable != nil && (zeroNil || (!hasFallback && m.sourceIsPointer(&src, field))) {
			field, setter = &GoField{name: field.name, typ: bf.nillable.typ}, bf.nillableSetter
		}

		mapping, err := m.resolveField(&src, field)
		if err != nil {
			return err
		}
		if mapping == nil {
			continue
		}

		// so does a converter that hands back a pointer
		if setter == bf.setter && bf.nillable != nil && mapping.produces != nil && mapping.produces.Kind == PointerKind {
			setter = bf.nillableSetter
		}

		setters[mapping] = setter
		resolved = append(resolved, mapping)
	}

	canFail, err := checkCanFail(f, resolved)
	if err != nil {
		return err
	}

	returnType := f.output.ref()
	if f.hasError {
		returnType = jen.Parens(jen.List(returnType, jen.Id("error")))
	}

	result := jen.Id(target.param.name)
	if f.hasError {
		result = jen.List(result, jen.Nil())
	}

	body := jen.Empty()
	for _, input := range src.inputs {
		if input.typ.Kind == PointerKind {
			body.Add(jen.If(jen.Id(input.name).Op("==").Nil()).Block(jen.Return(result.Clone()))).Line()
		}
	}

	ctx := &FunctionContext{
		body:       body,
		targetType: f.output,
		canFail:    canFail,
	}

	for _, r := range resolved {
		value := r.generator(ctx, r.input)
		body.Add(jen.Id(target.param.name).Dot(setters[r]).Call(value)).Line()
	}

	function := jen.Func().Params(jen.Id("m").Id(mapperName)).Id(f.name).Params(inputParams(f)...).Add(returnType)
	function.Block(body, jen.Return(result)).Line()

	m.file.Add(function)

	return nil
}

// hasSource reports whether field has a directive or a source field of its name.
func (m *Generator) hasSource(f *GoFunction, field *GoField) bool {
	if f.options[field.name] != nil {
		return true
	}

	_, path, err := m.findField(f.inputs, m.sourceName(f, field.name))

	return err != nil || path != nil
}

// sourceIsPointer reports whether the source field a plain mapping of field would
// read is a pointer.
func (m *Generator) sourceIsPointer(f *GoFunction, field *GoField) bool {
	source := m.sourceName(f, field.name)
	if source == "." || source == "-" {
		return false
	}

	_, path, err := m.findField(f.inputs, source)

	return err == nil && path != nil && path.typ().Kind == PointerKind
}
