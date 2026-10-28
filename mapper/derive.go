package mapper

import (
	"fmt"
	"go/types"
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
	fromStruct, toStruct := structBehind(from), structBehind(to)
	if fromStruct == nil || toStruct == nil {
		return nil, false
	}

	// a type go vet won't let be copied, a protobuf message, is only handled by
	// pointer: the helper takes and returns one, and the graph adapts the rest
	if hasLock(fromStruct.Raw) || hasLock(toStruct.Raw) {
		if m.deriveStruct(fromStruct, toStruct, true) == nil {
			return nil, false
		}

		return m.buildGenerator(from, to)
	}

	mapping := m.deriveStruct(fromStruct, toStruct, false)
	if mapping == nil {
		return nil, false
	}

	gen := mapping.generator
	switch {
	case from.Kind == PointerKind && to.Kind == PointerKind:
		gen = mapPointerToPointer(toStruct, gen)
	case to.Kind == PointerKind:
		gen = addressOf(gen)
	case from.Kind == PointerKind:
		gen = mapPointerToValue(to, gen)
	}

	return gen, mapping.canFail
}

// structBehind is typ if it is a struct, or the struct it points to.
func structBehind(typ *GoType) *GoType {
	if typ.Kind == PointerKind && typ.Element != nil {
		typ = typ.Element
	}
	if typ.Kind != StructKind {
		return nil
	}

	return typ
}

// deriveStruct generates `func (m Impl) mapAToB(in A) B` for two named structs, or
// with byPointer `mapAToB(in *A) *B`, and registers it as their converter for the
// rest of the interface. The helper follows the interface's toggles, but takes no
// directives. It returns nil if a field cannot be mapped.
func (m *Generator) deriveStruct(from, to *GoType, byPointer bool) *TypeMapping {
	if m.iface == nil || m.noDerive > 0 || namedType(from) == nil || namedType(to) == nil ||
		len(from.Fields) == 0 || len(to.Fields) == 0 {
		return nil
	}

	in, out := from, to
	if byPointer {
		in, out = from.pointer(), to.pointer()
	}

	pair := TypePair{from: in, to: out}
	if m.derived.failed[pair.String()] {
		return nil
	}

	name := m.helperName(from, to)

	// registered up front, so that a struct referring to itself finds the helper
	mapping := &TypeMapping{types: pair, generator: ownMethodConverter(name, false)}
	prev := m.graph.add(mapping)

	fn := &GoFunction{
		name:          name,
		inputs:        []*GoParameter{{name: "in", typ: in}},
		output:        out,
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
		log.Printf("not deriving %s to %s: %v", in, out, err)

		return nil
	}

	m.scoped = append(m.scoped, scopedMapping{mapping: mapping, prev: prev, helper: name})
	m.derived.names[name] = true
	fn.hasError = mapping.canFail

	log.Printf("converting %s to %s field by field, with %s", in, out, name)
	m.derived.code = append(m.derived.code, m.renderMethod(m.implName, fn, resolved, mapping.canFail))

	return mapping
}

// hasLock reports whether a value of typ holds a lock, which go vet's copylocks
// check forbids copying: a type whose pointer has Lock() and Unlock(), such as
// sync.Mutex or the [0]sync.Mutex protobuf embeds in every message, or a struct or
// array containing one.
func hasLock(typ types.Type) bool {
	return lockIn(typ, map[types.Type]bool{})
}

func lockIn(typ types.Type, seen map[types.Type]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true

	if _, isPtr := typ.(*types.Pointer); !isPtr {
		methods := types.NewMethodSet(types.NewPointer(typ))
		if methods.Lookup(nil, "Lock") != nil && methods.Lookup(nil, "Unlock") != nil {
			return true
		}
	}

	switch u := typ.Underlying().(type) {
	case *types.Struct:
		for i := 0; i < u.NumFields(); i++ {
			if lockIn(u.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Array:
		return lockIn(u.Elem(), seen)
	}

	return false
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
