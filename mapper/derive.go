package mapper

import (
	"fmt"
	"log"

	"github.com/dave/jennifer/jen"
)

// derivedHelpers tracks the struct conversions derived for one interface. Their
// code is only added to the file once the interface is complete.
type derivedHelpers struct {
	names  map[string]bool
	failed map[string]bool
	code   []jen.Code
}

func newDerivedHelpers() derivedHelpers {
	return derivedHelpers{names: map[string]bool{}, failed: map[string]bool{}}
}

// deriveConversion maps a struct onto another struct field by field when nothing
// else converts them, through a helper method generated on the implementation.
// A pointer on either side is nil-safe: nil gives nil for a pointer target, the
// zero value otherwise. It returns nil if any target field cannot be mapped.
func (m *Generator) deriveConversion(from, to *GoType) (GeneratorFunc, bool) {
	switch {
	case from.Kind == StructKind && to.Kind == StructKind:
		return m.deriveStruct(from, to)
	case from.Kind == StructKind && to.Kind == PointerKind && to.Element.Kind == StructKind:
		gen, canFail := m.deriveStruct(from, to.Element)
		if gen != nil {
			return addressOf(gen), canFail
		}
	case from.Kind == PointerKind && from.Element.Kind == StructKind && to.Kind == StructKind:
		gen, canFail := m.deriveStruct(from.Element, to)
		if gen != nil {
			return mapPointerToValue(to, gen), canFail
		}
	}

	return nil, false
}

// deriveStruct generates `func (m Impl) mapAToB(in A) B` for two named structs and
// registers it as their converter for the rest of the interface. The helper follows
// the interface's toggles, but takes no directives.
func (m *Generator) deriveStruct(from, to *GoType) (GeneratorFunc, bool) {
	if m.iface == nil || m.noDerive > 0 || namedType(from) == nil || namedType(to) == nil ||
		len(from.Fields) == 0 || len(to.Fields) == 0 {
		return nil, false
	}

	pair := TypePair{from: from, to: to}
	if m.derived.failed[pair.String()] {
		return nil, false
	}

	name := m.helperName(from, to)

	// registered up front, so that a struct referring to itself finds the helper
	mapping := &TypeMapping{types: pair, generator: ownMethodConverter(name, false)}
	prev := m.graph.add(mapping)

	fn := &GoFunction{
		name:          name,
		inputs:        []*GoParameter{{name: "in", typ: from}},
		output:        to,
		options:       map[string]*FieldOptions{},
		skipUnmatched: m.iface.skipUnmatched,
	}

	// what an attempt that is given up on leaves behind is taken back with it:
	// enum problems, and helpers derived on the way, which may call this one
	problemsBefore, scopedBefore, codeBefore := len(m.enumProblems), len(m.scoped), len(m.derived.code)

	resolved, err := m.buildMapping(fn, to.Fields)
	if err == nil && anyCanFail(resolved) {
		// resolved again, a field referring back to the helper has to handle its error
		mapping.canFail = true
		mapping.generator = ownMethodConverter(name, true)
		resolved, err = m.buildMapping(fn, to.Fields)
	}

	if err != nil {
		m.enumProblems = m.enumProblems[:problemsBefore]
		m.unscope(scopedBefore)
		m.derived.code = m.derived.code[:codeBefore]
		m.graph.restore(mapping, prev)
		m.derived.failed[pair.String()] = true
		log.Printf("not deriving %s to %s: %v", from, to, err)

		return nil, false
	}

	m.scoped = append(m.scoped, scopedMapping{mapping: mapping, prev: prev, helper: name})
	m.derived.names[name] = true
	fn.hasError = mapping.canFail

	log.Printf("converting %s to %s field by field, with %s", from, to, name)
	m.derived.code = append(m.derived.code, m.renderMethod(m.implName, fn, resolved, mapping.canFail))

	return mapping.generator, mapping.canFail
}

// helperName names the helper for a derived conversion, unique within the interface.
func (m *Generator) helperName(from, to *GoType) string {
	base := "map" + from.Name + "To" + to.Name
	name := base
	for i := 2; m.derived.names[name]; i++ {
		name = fmt.Sprintf("%s%d", base, i)
	}

	return name
}

// addressOf takes the address of what gen makes.
func addressOf(gen GeneratorFunc) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		return mapToPointer(identityMapping)(ctx, gen(ctx, input))
	}
}
