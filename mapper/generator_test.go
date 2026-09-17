package mapper

import (
	"go/token"
	"go/types"
	"testing"
)

func named(pkg *types.Package, name string, underlying types.Type) *types.Named {
	return types.NewNamed(types.NewTypeName(token.NoPos, pkg, name, nil), underlying, nil)
}

func TestIsDirectConversion(t *testing.T) {
	entPkg := types.NewPackage("example.com/ent/galaxy", "galaxy")
	storePkg := types.NewPackage("example.com/store", "store")

	entState := named(entPkg, "State", types.Typ[types.String])
	storeState := named(storePkg, "GalaxyState", types.Typ[types.String])

	structFields := []*types.Var{types.NewField(token.NoPos, nil, "Name", types.Typ[types.String], false)}
	entStruct := named(entPkg, "Point", types.NewStruct(structFields, nil))
	storeStruct := named(storePkg, "Point", types.NewStruct(structFields, nil))

	entCount := named(entPkg, "Count", types.Typ[types.Int64])

	tests := []struct {
		name string
		from types.Type
		to   types.Type
		want bool
	}{
		{"named string to named string", entState, storeState, true},
		{"named string to named string, reversed", storeState, entState, true},
		{"unnamed string to named string", types.Typ[types.String], storeState, true},
		{"named string to unnamed string", entState, types.Typ[types.String], true},
		{"identical named types", entState, entState, false},
		{"identical basic types", types.Typ[types.String], types.Typ[types.String], false},
		{"named struct to named struct", entStruct, storeStruct, true},
		{"int64 to int32", types.Typ[types.Int64], types.Typ[types.Int32], true},
		{"named int64 to int32", entCount, types.Typ[types.Int32], true},
		{"int to named uint8", types.Typ[types.Int], named(storePkg, "State", types.Typ[types.Uint8]), true},
		{"int to float64", types.Typ[types.Int], types.Typ[types.Float64], false},
		{"string to byte slice", types.Typ[types.String], types.NewSlice(types.Typ[types.Byte]), true},
		{"byte slice to named string", types.NewSlice(types.Typ[types.Uint8]), entState, true},
		{"named byte slice to string", named(entPkg, "Name", types.NewSlice(types.Typ[types.Byte])), types.Typ[types.String], true},
		{"string to int slice", types.Typ[types.String], types.NewSlice(types.Typ[types.Int]), false},
		{"named string to named int", entState, named(storePkg, "Code", types.Typ[types.Int]), false},
		{"pointers to convertible types", types.NewPointer(entState), types.NewPointer(storeState), false},
		{"unrelated named structs", entStruct, entCount, false},
		{"nil types", nil, storeState, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDirectConversion(tt.from, tt.to); got != tt.want {
				t.Errorf("isDirectConversion(%v, %v) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
