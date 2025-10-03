package mapper

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"
	"log"
	"os"
	"strings"

	"github.com/dave/jennifer/jen"
)

var typeMappingRegistry []*TypeMapping

func registerTypeMappings(mapping ...*TypeMapping) {
	typeMappingRegistry = append(typeMappingRegistry, mapping...)
}

func NewGenerator(pkg string) *Generator {
	gen := &Generator{
		file:     jen.NewFile(pkg),
		mappings: map[string]*TypeMapping{},
		graph:    NewTypeGraph(),
	}

	for _, mapper := range typeMappingRegistry {
		gen.AddTypeMapping(mapper)
	}

	return gen
}

type Generator struct {
	file     *jen.File
	mappings map[string]*TypeMapping
	graph    *TypeGraph
}

func (m *Generator) AddTypeMapping(mapping *TypeMapping) {
	name := mapping.types.String()
	m.mappings[name] = mapping
	m.graph.add(mapping)
}

func (m *Generator) GenerateInterface(in *GoInterface) error {
	name := in.Name + "Impl"
	mapperStruct := jen.Type().Id(name).Struct()
	typeCheck := jen.Var().Id("_").Qual(in.Pkg, in.Name).Op("=").Parens(jen.Op("*").Id(name)).Parens(jen.Nil())

	// add type mappings
	for _, method := range in.Methods {
		// only handle cases with one input for now
		if len(method.inputs) == 1 {
			m.AddTypeMapping(&TypeMapping{
				types: TypePair{
					from: method.inputs[0].typ,
					to:   method.output,
				},
				generator: ownMethodConverter(method.name, method.hasError),
				canFail:   method.hasError,
			})
		}

		// discover protobuf enums
		m.discoverEnum(method.output)
		for _, input := range method.inputs {
			m.discoverEnum(input.typ)
		}
	}

	for _, method := range in.Methods {
		err := m.generateMethod(name, method)
		if err != nil {
			log.Println(method.name, err)
			//return err
		}
	}

	m.file.Add(mapperStruct)
	m.file.Add(typeCheck)

	return nil
}

func (m *Generator) discoverEnum(typ *GoType) {
	if typ.IsProtoEnum {
		registerProtoEnum(m, typ)
	}

	for _, field := range typ.Fields {
		m.discoverEnum(field.typ)
	}

	if typ.Element != nil {
		m.discoverEnum(typ.Element)
	}
}

func (m *Generator) findField(inputs []*GoParameter, source string) (*GoParameter, GoFieldPath) {
	for _, input := range inputs {

		// see if the input contains our output, otherwise continue
		if inputField := input.typ.findField(source); inputField != nil {
			return input, inputField
		}
	}

	return nil, nil
}

func (m *Generator) buildMapping(f *GoFunction) ([]*ResolvedMapping, error) {
	// look at all the Fields required for the output
	// and try to map them to input Fields

	fields := f.output.Fields
	if f.output.Element != nil {
		fields = f.output.Element.Fields
	}

	var mappings []*ResolvedMapping
	for _, field := range fields {
		source := field.name

		// apply output options
		if options, ok := f.options[field.name]; ok {
			if options.from == "-" {
				continue
			} else if options.from != "" {
				source = options.from
			}
		}

		// resolve the output
		var mapping *ResolvedMapping

		if source == "." {
			// special case of referring directly to the input type
			for _, input := range f.inputs {
				generator, canFail := m.buildGenerator(input.typ, field.typ)

				if generator != nil {
					mapping = &ResolvedMapping{
						output:    field,
						input:     jen.Id(input.name),
						generator: generator,
						canFail:   canFail,
					}
				}
			}

		} else if param, inputField := m.findField(f.inputs, source); inputField != nil {
			// see if the input contains our output, otherwise continue

			generator, canFail := m.buildGenerator(inputField.typ(), field.typ)
			if generator != nil {
				mapping = &ResolvedMapping{
					output:    field,
					input:     jen.Id(param.name).Op(".").Add(inputField.id()),
					generator: generator,
					canFail:   canFail,
				}
			} else {
				log.Printf("Can't convert from type %s to %s", inputField.typ(), field.typ)
			}
		}

		if mapping == nil {
			// If no custom mapper was used and no direct mapping found, try buildGenerator
			for _, input := range f.inputs {
				generator, canFail := m.buildGenerator(input.typ, field.typ)
				if generator != nil {
					mapping = &ResolvedMapping{
						output:    field,
						input:     jen.Id(input.name),
						generator: generator,
						canFail:   canFail,
					}
					break
				}
			}
		}

		if mapping == nil {
			return nil, fmt.Errorf("could not find source for output %s", field.name)
		}

		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

func (m *Generator) buildGenerator(from *GoType, to *GoType) (GeneratorFunc, bool) {
	// build a type pair and see if we can map the fields
	pair := TypePair{
		from: from,
		to:   to,
	}

	// as a special case same types can be assigned without modification
	if types.AssignableTo(pair.from.Raw, pair.to.Raw) {
		return identityMapping, false
	}

	if pair.from.Kind == PointerKind {
		if types.AssignableTo(pair.from.Element.Raw, pair.to.Raw) {
			return mapFromPointer(identityMapping, pair.to), false
		}
	}

	if pair.to.Kind == PointerKind {
		if types.AssignableTo(pair.from.Raw, pair.to.Element.Raw) {
			return mapToPointer(identityMapping), false
		}
	}

	// otherwise, see if we can map the fields
	mappers := m.graph.getMapping(pair.from, pair.to)
	if mappers != nil {
		gen, canFail := compose(pair, mappers)

		var names []string
		for _, mapper := range mappers {
			names = append(names, fmt.Sprintf("%s:%s", mapper.types.from.Name, mapper.types.to.Name))
		}
		log.Printf("converting %s to %s using %d converters: %s", pair.from.String(), pair.to.String(), len(mappers), strings.Join(names, " -> "))

		return gen, canFail
	}

	// if we have no direct match, see if the potential element types can be matched
	if from.Kind == SliceKind && to.Kind == SliceKind {
		gen, canFail := m.buildGenerator(from.Element, to.Element)
		if gen != nil {
			return sliceMapping(to.Element, gen), canFail
		}
	}

	if from.Kind != to.Kind {

	}

	return nil, false
}

func (m *Generator) generateMethod(mapperName string, f *GoFunction) error {
	to := f.output
	if to.Kind != StructKind && to.Kind != PointerKind && to.Element.Kind != StructKind {
		return errors.New("target type must be a struct or a pointer to a struct")
	}

	resolved, err := m.buildMapping(f)
	if err != nil {
		return err
	}

	canFail := false
	for _, r := range resolved {
		if r.canFail {
			canFail = true
			break
		}
	}

	if canFail && !f.hasError {
		return errors.New("at least one converter can fail, function needs to return error")
	} else if !canFail && f.hasError {
		return errors.New("no converter can fail, function can never return error")
	}

	returnType := jen.Empty()
	if to.Kind == PointerKind {
		returnType.Op("*")
	}
	returnType.Qual(to.Pkg, to.Name)

	if f.hasError {
		returnType = jen.Parens(jen.List(returnType, jen.Id("error")))
	}

	var inputTypes []jen.Code
	for i, input := range f.inputs {
		name := input.name
		if name == "" {
			name = fmt.Sprintf("input%d", i)
		}
		inputType := jen.Id(name)

		if input.typ.Kind == PointerKind {
			inputType.Op("*")
		}
		inputType.Qual(input.typ.Pkg, input.typ.Name)

		inputTypes = append(inputTypes, inputType)
	}

	function := jen.Func().Params(jen.Id("m").Id(mapperName)).Id(f.name).Params(jen.List(inputTypes...)).Add(returnType)

	// prepare all the Fields
	body := jen.Empty()
	ctx := &FunctionContext{
		body:       body,
		targetType: f.output,
		canFail:    canFail,
	}

	var assignments []jen.Code
	for _, r := range resolved {
		assignments = append(assignments, jen.Line().Id(r.output.name).Op(":").Add(r.generator(ctx, r.input)))
	}

	assignments = append(assignments, jen.Line())

	// initialize the struct
	returnValue := jen.Empty()
	if to.Kind == PointerKind {
		returnValue.Op("&")
	}

	returnValue.Qual(to.Pkg, to.Name).Values(assignments...)

	if f.hasError {
		returnValue = jen.List(returnValue, jen.Nil())
	}

	function.Block(body, jen.Return(returnValue)).Line()

	m.file.Add(function)

	return nil
}

func (m *Generator) String() (string, error) {
	buf := new(bytes.Buffer)
	err := m.file.Render(buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m *Generator) Render() {
	log.Println(m.file.Render(os.Stdout))
}

func compose(pair TypePair, mappings []*TypeMapping) (GeneratorFunc, bool) {
	canFail := false

	for _, mapping := range mappings {
		if mapping.canFail {
			canFail = true
			break
		}
	}

	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		output := input

		outputType := pair.from
		for _, mapping := range mappings {
			inputType := mapping.types.from
			generator := mapping.generator

			if (outputType.Kind == PointerKind) != (inputType.Kind == PointerKind) {
				// need to adjust for pointers

				if inputType.Kind == PointerKind {
					generator = mapToPointer(generator)
				} else {
					generator = mapFromPointer(generator, inputType)
				}
			}

			output = generator(ctx, output)
			outputType = mapping.types.to
		}

		return output
	}, canFail
}

func mapToPointer(gen GeneratorFunc) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		tmpVar := ctx.tmpVar()

		ctx.add(jen.Id(tmpVar).Op(":=").Add(input).Line())

		return gen(ctx, jen.Op("&").Id(tmpVar))
	}
}

func mapFromPointer(gen GeneratorFunc, typ *GoType) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		tmpVar := ctx.tmpVar()

		ctx.add(jen.Var().Id(tmpVar).Add(typ.qual()).Line())
		ctx.add(jen.If(input).Op("!=").Nil().Block(
			jen.Id(tmpVar).Op("=").Op("*").Parens(input),
		).Line().Line())

		return gen(ctx, jen.Id(tmpVar))
	}
}

func sliceMapping(targetType *GoType, gen GeneratorFunc) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		tmpVar := ctx.tmpVar()

		initializer := jen.Id(tmpVar).Op(":=").Make(jen.List(jen.Op("[]").Add(targetType.qual()), jen.Len(input)))

		subCtx := &FunctionContext{
			targetType: ctx.targetType,
			body:       jen.Empty(),
			canFail:    false,
			varCounter: 0,
		}

		value := gen(subCtx, jen.Add(input).Index(jen.Id("i")))

		loopBody := jen.Id(tmpVar).Index(jen.Id("i")).Op("=").Add(value)
		loop := jen.For(jen.Id("i").Op(":=").Range().Add(input)).Block(subCtx.body.Add(loopBody))

		ctx.add(
			initializer.Line(),
			loop.Line().Line(),
		)

		return jen.Id(tmpVar)
	}
}

func identityMapping(ctx *FunctionContext, input jen.Code) jen.Code {
	return input
}

func ownMethodConverter(name string, hasError bool) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		call := jen.Id("m").Dot(name).Call(input)

		if hasError {
			return ctx.handleError(call)
		}

		return call
	}
}
