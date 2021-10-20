package mapper

import (
	"go/types"

	"github.com/dave/jennifer/jen"
)

func init() {
	var builtinMappers []*TypeMapping

	intTypes := []types.BasicKind{types.Int, types.Int8, types.Int16, types.Int32, types.Int64}
	uIntTypes := []types.BasicKind{types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64}

	var allIntTypes []types.BasicKind
	allIntTypes = append(allIntTypes, intTypes...)
	allIntTypes = append(allIntTypes, uIntTypes...)

	// integer formatting
	builtinMappers = append(builtinMappers, integerFormatter(types.Int64, "FormatInt"))
	builtinMappers = append(builtinMappers, integerFormatter(types.Uint64, "FormatUint"))

	// integer type conversion
	for i, intType := range allIntTypes {
		input := getGoType(types.Typ[intType])

		for j := 0; j < len(allIntTypes); j++ {
			if i == j {
				continue
			}

			output := getGoType(types.Typ[allIntTypes[j]])

			builtinMappers = append(builtinMappers, &TypeMapping{
				types: TypePair{
					from: input,
					to:   output,
				},
				generator: typeConverter(output.Name),
			})
		}
	}

	// int parsing
	for _, intType := range intTypes {
		builtinMappers = append(builtinMappers, integerParser(intType, "ParseInt"))
	}

	// uint parsing
	for _, uIntType := range uIntTypes {
		builtinMappers = append(builtinMappers, integerParser(uIntType, "ParseUint"))
	}

	registerTypeMappings(builtinMappers...)
}

func integerFormatter(typ types.BasicKind, name string) *TypeMapping {
	return &TypeMapping{
		types: TypePair{
			from: getGoType(types.Typ[typ]),
			to:   StringType,
		},
		generator: mapNumericToString(name),
	}
}

func integerParser(typ types.BasicKind, parser string) *TypeMapping {
	sizes := types.SizesFor("gc", "amd64")
	intType := types.Typ[typ]

	return &TypeMapping{
		types: TypePair{
			from: StringType,
			to:   getGoType(intType),
		},
		generator: mapStringToNumeric(intType, parser, int(sizes.Sizeof(intType)*8)),
		canFail:   true,
	}
}

func typeConverter(to string) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		return jen.Id(to).Parens(input)
	}
}

func functionConverter(pkg, function string, hasError bool) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		fn := jen.Qual(pkg, function)

		if hasError {
			return ctx.handleError(fn.Call(input))
		}

		return fn.Call(input)
	}
}

func fieldConverter(name string) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		return jen.Add(input).Dot(name)
	}
}

func methodConverter(name string, typ *GoType, hasError bool) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		call := jen.Add(input).Dot(name).Call()

		if hasError {
			return ctx.handleError(call)
		}

		return call
	}
}

func mapNumericToString(formatter string) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		return jen.Qual("strconv", formatter).Call(input, jen.Lit(10))
	}
}

func mapStringToNumeric(typ *types.Basic, parser string, bitSize int) GeneratorFunc {
	return func(ctx *FunctionContext, input jen.Code) jen.Code {
		tmp := ctx.handleError(jen.Qual("strconv", parser).
			Call(input, jen.Lit(10), jen.Lit(bitSize)))

		if typ.Kind() != types.Int64 && typ.Kind() != types.Uint64 {
			tmp = jen.Id(typ.Name()).Parens(tmp)
		}

		return tmp
	}
}
