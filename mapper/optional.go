package mapper

import (
	"errors"
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

// buildFieldGenerator converts a source of type src into field, as buildGenerator
// does, honouring the field's `default:` and `zero:"nil"` directives.
func (m *Generator) buildFieldGenerator(f *GoFunction, field *GoField, src *GoType) (GeneratorFunc, bool, error) {
	fallback, err := m.fallbackValue(f, field, src)
	if err != nil {
		return nil, false, err
	}
	if fallback != nil {
		gen, canFail := m.buildGenerator(src.Element, field.typ)
		if gen == nil {
			return nil, false, nil
		}

		return mapPointerToValueOr(field.typ, fallback, gen, true), canFail, nil
	}

	gen, canFail := m.buildGenerator(src, field.typ)
	if gen == nil {
		return nil, false, nil
	}

	gen, err = zeroToNil(f, field, src, field.typ, field.typ, gen)

	return gen, canFail, err
}

// rejectOptional fails for a field with a `default:` or `zero:"nil"` directive,
// for where they do not apply: element by element, or on an input that is never nil.
func rejectOptional(f *GoFunction, field *GoField) error {
	if options := f.options[field.name]; options != nil && (options.fallback != "" || options.zeroNil) {
		return errors.New(`default: and zero:"nil" apply to a single source field, not element by element or to the input itself`)
	}

	return nil
}

// fallbackValue is the code for a field's `default:` directive, nil without one. The
// value is a constant of the field's type, by name or value, or a literal of it.
func (m *Generator) fallbackValue(f *GoFunction, field *GoField, src *GoType) (jen.Code, error) {
	options := f.options[field.name]
	if options == nil || options.fallback == "" {
		return nil, nil
	}

	if field.typ.Kind == PointerKind {
		return nil, errors.New("default: is for a field that is no pointer, a pointer field keeps nil")
	}
	if src.Kind != PointerKind {
		return nil, fmt.Errorf("default: the source is a %s, which is never nil", src)
	}

	code, _, err := m.resolveEnumRef(field.typ, options.fallback, "default:")

	return code, err
}

// zeroToNil wraps gen, which turns a src into what produces is, so that the zero
// value of src gives nil for the pointer field dst when the field has a
// `zero:"nil"` directive. A value gen produces is addressed.
func zeroToNil(f *GoFunction, field *GoField, src, dst, produces *GoType, gen GeneratorFunc) (GeneratorFunc, error) {
	options := f.options[field.name]
	if options == nil || !options.zeroNil {
		return gen, nil
	}

	if dst.Kind != PointerKind {
		return nil, errors.New(`zero:"nil" is for a pointer field`)
	}
	if src.Kind == PointerKind {
		return nil, fmt.Errorf(`zero:"nil": the source is the pointer %s already`, src)
	}

	notZero, err := notZeroCheck(src)
	if err != nil {
		return nil, err
	}

	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		outVar := ctx.tmpVar()
		ctx.add(jen.Var().Id(outVar).Add(dst.ref()).Line())

		subCtx := &FunctionContext{
			targetType: ctx.targetType,
			body:       jen.Empty(),
			varCounter: ctx.varCounter,
		}
		converted := gen(subCtx, input)
		ctx.varCounter = subCtx.varCounter

		assign := jen.Id(outVar).Op("=").Add(converted)
		if produces != nil && produces.Kind != PointerKind {
			valueVar := ctx.tmpVar()
			assign = jen.Id(valueVar).Op(":=").Add(converted).Line().Id(outVar).Op("=").Op("&").Id(valueVar)
		}

		ctx.add(jen.If(notZero(input)).Block(jen.Empty().Add(subCtx.body).Add(assign)).Line())

		return jen.Id(outVar)
	}, nil
}

// notZeroCheck renders the condition that a value of typ is not its zero value:
// `!v.IsZero()` for a type with that method, as time.Time has, else a comparison.
func notZeroCheck(typ *GoType) (func(jen.Code) jen.Code, error) {
	if types.Implements(typ.Raw, IsZeroerInterface) {
		return func(v jen.Code) jen.Code { return jen.Op("!").Add(v).Dot("IsZero").Call() }, nil
	}

	if !types.Comparable(typ.Raw) {
		return nil, fmt.Errorf(`zero:"nil": %s has no IsZero() and cannot be compared to its zero value`, typ)
	}

	zero := typ.zero()
	if typ.Kind == StructKind || typ.Kind == ArrayKind {
		// a composite literal in an if condition needs parentheses
		zero = jen.Parens(typ.ref().Values())
	}

	return func(v jen.Code) jen.Code { return jen.Add(v).Op("!=").Add(zero) }, nil
}

// mapPointerToValueOr is mapPointerToValue with fallback instead of the zero value
// for a nil input. Without deref, gen is handed the pointer rather than its value.
func mapPointerToValueOr(target *GoType, fallback jen.Code, gen GeneratorFunc, deref bool) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		outVar := ctx.tmpVar()
		inVar := ctx.tmpVar()

		ctx.add(jen.Var().Id(outVar).Add(target.ref()).Op("=").Add(fallback).Line())

		subCtx := &FunctionContext{
			targetType: ctx.targetType,
			body:       jen.Empty(),
			varCounter: ctx.varCounter,
		}
		value := jen.Id(inVar)
		if deref {
			derefVar := subCtx.tmpVar()
			subCtx.add(jen.Id(derefVar).Op(":=").Op("*").Id(inVar).Line())
			value = jen.Id(derefVar)
		}
		converted := gen(subCtx, value)
		ctx.varCounter = subCtx.varCounter

		ctx.add(jen.If(
			jen.Id(inVar).Op(":=").Add(input),
			jen.Id(inVar).Op("!=").Nil(),
		).Block(jen.Empty().Add(subCtx.body).Add(jen.Id(outVar).Op("=").Add(converted))).Line())

		return jen.Id(outVar)
	}
}
