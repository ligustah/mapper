package mapper

import (
	"fmt"
	"time"

	"github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// the module being generated may not use protobuf at all
func init() {
	registerOptionalMapper[*timestamppb.Timestamp, time.Time]((*timestamppb.Timestamp).AsTime)
	registerOptionalMapper[time.Time, *timestamppb.Timestamp](timestamppb.New)

	registerOptionalMapper[*durationpb.Duration, time.Duration]((*durationpb.Duration).AsDuration)
	registerOptionalMapper[time.Duration, *durationpb.Duration](durationpb.New)
}

func registerProtoEnum(generator *Generator, enum *GoType) {
	// use the base enum type for name/value lookups
	baseEnum := enum
	if enum.Kind == PointerKind && enum.Element != nil {
		baseEnum = enum.Element
	}

	name := fmt.Sprintf("%s_name", baseEnum.Name)
	value := fmt.Sprintf("%s_value", baseEnum.Name)

	generator.AddTypeMapping(
		&TypeMapping{
			types: TypePair{
				from: enum,
				to:   StringType,
			},
			generator: func(ctx *FunctionContext, input jen.Code) jen.Code {
				tmp := ctx.tmpVar()

				ctx.add(jen.Var().Id(tmp).Qual("", "string").Line())

				// dereference pointer if needed
				inputExpr := input
				if enum.Kind == PointerKind {
					inputExpr = jen.Op("*").Parens(input)
				}

				ctx.add(jen.IfFunc(func(group *jen.Group) {
					group.List(jen.Id("str"), jen.Id("ok")).Op(":=").Qual(baseEnum.Pkg, name).Index(jen.Qual("", "int32").Parens(inputExpr))
					group.Id("ok")
				}).BlockFunc(func(group *jen.Group) {
					group.Id(tmp).Op("=").Id("str")
				}).Else().BlockFunc(func(group *jen.Group) {
					fmtString := jen.Lit(fmt.Sprintf("invalid integer value for enum %s: %%d", baseEnum.Name))

					group.Return(ctx.targetType.zero(), jen.Qual("fmt", "Errorf").Call(fmtString, inputExpr))
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
					group.List(jen.Id("val"), jen.Id("ok")).Op(":=").Qual(baseEnum.Pkg, value).Index(input)
					group.Id("ok")
				}).BlockFunc(func(group *jen.Group) {
					group.Id(tmp).Op("=").Add(enum.qual()).Parens(jen.Id("val"))
				}).Else().BlockFunc(func(group *jen.Group) {
					fmtString := jen.Lit(fmt.Sprintf("invalid string value for enum %s: %%s", baseEnum.Name))

					group.Return(ctx.targetType.zero(), jen.Qual("fmt", "Errorf").Call(fmtString, input))
				}).Line().Line())

				return jen.Id(tmp)
			},
			canFail: true,
		},
	)
}
