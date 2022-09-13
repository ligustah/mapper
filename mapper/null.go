package mapper

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

func init() {
	var nullMappers []*TypeMapping

	type m struct {
		name string
		typ  *GoType
	}

	nullTypes := []m{
		{"String", StringType},
		{"Time", MustLookupType("time.Time")},
	}

	for _, nullType := range nullTypes {
		n := MustLookupType(fmt.Sprintf("github.com/volatiletech/null.%s", nullType.name))

		nullMappers = append(nullMappers, &TypeMapping{
			types: TypePair{
				from: n,
				to:   nullType.typ,
			},
			generator: fieldConverter(nullType.name),
		})

		nullMappers = append(nullMappers, &TypeMapping{
			types: TypePair{
				from: n,
				to:   nullType.typ.pointer(),
			},
			generator: methodConverter("Ptr", n, false),
		})

		nullMappers = append(nullMappers, &TypeMapping{
			types: TypePair{
				from: nullType.typ,
				to:   n,
			},
			generator: func(typ m) GeneratorFunc {
				return func(ctx *FunctionContext, input jen.Code) jen.Code {
					f := functionConverter("github.com/volatiletech/null", fmt.Sprintf("New%s", typ.name), false)

					tmp := ctx.tmpVar()
					ctx.add(jen.Id(tmp).Op(":=").Add(input).Line())

					isValid := jen.Op("!").Parens(jen.Id(tmp).Add(typ.typ.isZero()))

					return f(ctx, jen.List(jen.Id(tmp), isValid))
				}
			}(nullType),
		})

		nullMappers = append(nullMappers, &TypeMapping{
			types: TypePair{
				from: nullType.typ.pointer(),
				to:   n,
			},
			generator: functionConverter("github.com/volatiletech/null", fmt.Sprintf("%sFromPtr", nullType.name), false),
		})
	}

	registerTypeMappings(nullMappers...)
}
