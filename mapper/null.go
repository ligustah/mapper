package mapper

import "fmt"

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
			generator: functionConverter("github.com/volatiletech/null", fmt.Sprintf("%sFrom", nullType.name), false),
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
