package mapper

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

func init() {
	var protoMappers []*TypeMapping

	goTime := MustLookupType("time.Time")
	pbTime := MustLookupType("*google.golang.org/protobuf/types/known/timestamppb.Timestamp")

	protoMappers = append(protoMappers, &TypeMapping{
		types: TypePair{
			from: goTime,
			to:   pbTime,
		},
		generator: functionConverter("google.golang.org/protobuf/types/known/timestamppb", "New", false),
	})

	protoMappers = append(protoMappers, &TypeMapping{
		types: TypePair{
			from: pbTime,
			to:   goTime,
		},
		generator: methodConverter("AsTime", pbTime, false),
	})

	registerTypeMappings(protoMappers...)
}

func registerProtoEnum(generator *Generator, enum *GoType) {
	name := fmt.Sprintf("%s_name", enum.Name)
	value := fmt.Sprintf("%s_value", enum.Name)

	generator.AddTypeMapping(
		&TypeMapping{
			types: TypePair{
				from: enum,
				to:   StringType,
			},
			generator: func(ctx *FunctionContext, input jen.Code) jen.Code {
				tmp := ctx.tmpVar()

				ctx.add(jen.Var().Id(tmp).Qual("", "string").Line())

				ctx.add(jen.IfFunc(func(group *jen.Group) {
					group.List(jen.Id("str"), jen.Id("ok")).Op(":=").Qual(enum.Pkg, name).Index(jen.Qual("", "int32").Parens(input))
					group.Id("ok")
				}).BlockFunc(func(group *jen.Group) {
					group.Id(tmp).Op("=").Id("str")
				}).Else().BlockFunc(func(group *jen.Group) {
					fmtString := jen.Lit(fmt.Sprintf("invalid string value for enum %s: %%d", enum.Name))

					group.Return(ctx.targetType.zero(), jen.Qual("fmt", "Errorf").Call(fmtString, input))
				}).Line().Line())

				return jen.Id(tmp)
			},
			canFail: true,
		})

	generator.AddTypeMapping(
		&TypeMapping{
			types: TypePair{
				from: StringType,
				to:   enum,
			},
			generator: func(ctx *FunctionContext, input jen.Code) jen.Code {
				tmp := ctx.tmpVar()

				ctx.add(jen.Var().Id(tmp).Add(enum.qual()).Line())

				ctx.add(jen.IfFunc(func(group *jen.Group) {
					group.List(jen.Id("val"), jen.Id("ok")).Op(":=").Qual(enum.Pkg, value).Index(input)
					group.Id("ok")
				}).BlockFunc(func(group *jen.Group) {
					group.Id(tmp).Op("=").Add(enum.qual()).Parens(jen.Id("val"))
				}).Else().BlockFunc(func(group *jen.Group) {
					fmtString := jen.Lit(fmt.Sprintf("invalid integer value for enum %s: %%d", enum.Name))

					group.Return(ctx.targetType.zero(), jen.Qual("fmt", "Errorf").Call(fmtString, input))
				}).Line().Line())

				return jen.Id(tmp)
			},
			canFail: true,
		},
	)
}
