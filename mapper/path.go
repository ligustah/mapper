package mapper

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

// guardPath wraps gen so that reading path off param cannot dereference a nil
// pointer. A pointer segment before the leaf is read through its `<Name>OrErr()`
// accessor when the enclosing type declares one and the method returns an error,
// which is how ent reports an edge that was not loaded. Otherwise the leaf is read
// behind a nil check and stays the zero value while a segment is nil. The bool
// reports whether the guard can fail.
func guardPath(f *GoFunction, param *GoParameter, path GoFieldPath, gen GeneratorFunc) (GeneratorFunc, bool) {
	orErr := make([]bool, len(path))
	guarded, canFail, nilChecked := false, false, false

	parent := param.typ
	for i, field := range path {
		if i < len(path)-1 && field.typ.Kind == PointerKind {
			guarded = true

			// once a segment is nil-checked the rest of the path sits inside that
			// check, an accessor call cannot be interleaved with it
			if f.hasError && !nilChecked && hasOrErr(parent, field) {
				orErr[i] = true
				canFail = true
			} else {
				nilChecked = true
			}
		}
		parent = field.typ
	}

	if !guarded {
		return gen, false
	}

	return func(ctx *FunctionContext, _ jen.Code) jen.Code {
		cur := jen.Id(param.name)
		var conds []jen.Code

		for i, field := range path {
			if orErr[i] {
				cur = ctx.handleError(cur.Clone().Dot(field.name + "OrErr").Call())
				continue
			}

			cur = cur.Clone().Dot(field.name)
			if i < len(path)-1 && field.typ.Kind == PointerKind {
				conds = append(conds, cur.Clone().Op("!=").Nil())
			}
		}

		if len(conds) == 0 {
			return gen(ctx, cur)
		}

		cond := conds[0]
		for _, c := range conds[1:] {
			cond = jen.Add(cond).Op("&&").Add(c)
		}

		tmp := ctx.tmpVar()
		ctx.add(jen.Var().Id(tmp).Add(path.typ().ref()).Line())
		ctx.add(jen.If(cond).Block(jen.Id(tmp).Op("=").Add(cur)).Line().Line())

		return gen(ctx, jen.Id(tmp))
	}, canFail
}

// hasOrErr reports whether the type holding field declares `<Field>OrErr() (T, error)`
// with T the field's own type, the accessor ent generates for every edge.
func hasOrErr(parent *GoType, field *GoField) bool {
	named := namedType(parent)
	if named == nil {
		return false
	}

	for i := 0; i < named.NumMethods(); i++ {
		method := named.Method(i)
		if method.Name() != field.name+"OrErr" {
			continue
		}

		sig := method.Type().(*types.Signature)
		if sig.Params().Len() != 0 || sig.Results().Len() != 2 {
			return false
		}

		return isErrorType(sig.Results().At(1).Type()) &&
			types.Identical(sig.Results().At(0).Type(), field.typ.Raw)
	}

	return false
}

// namedType is the named type behind typ, looking through one pointer.
func namedType(typ *GoType) *types.Named {
	raw := typ.Raw
	if ptr, ok := raw.(*types.Pointer); ok {
		raw = ptr.Elem()
	}

	named, _ := raw.(*types.Named)

	return named
}

func isErrorType(typ types.Type) bool {
	return types.Identical(typ, types.Universe.Lookup("error").Type())
}
