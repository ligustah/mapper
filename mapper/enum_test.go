package mapper

import (
	"go/constant"
	"go/token"
	"go/types"
	"testing"
)

type constDef struct {
	name  string
	value constant.Value
}

// enumType builds a named type together with the constants declared for it, in a
// package of its own, mirroring how an enum shows up in a loaded package.
func enumType(pkgPath, typeName string, underlying types.Type, consts ...constDef) *types.Named {
	pkg := types.NewPackage(pkgPath, typeName)
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, typeName, nil), underlying, nil)
	pkg.Scope().Insert(named.Obj())

	for _, c := range consts {
		pkg.Scope().Insert(types.NewConst(token.NoPos, pkg, c.name, named, c.value))
	}

	return named
}

func str(s string) constant.Value { return constant.MakeString(s) }
func num(i int64) constant.Value  { return constant.MakeInt64(i) }

func TestCheckEnumConversion(t *testing.T) {
	tests := []struct {
		name string
		from types.Type
		to   types.Type
		want int
	}{
		{
			name: "matching string enums",
			from: enumType("ent/galaxy", "State", types.Typ[types.String],
				constDef{"StateActive", str("active")}, constDef{"StateRetired", str("retired")}),
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateActive", str("active")}, constDef{"GalaxyStateRetired", str("retired")}),
			want: 0,
		},
		{
			name: "wider target is fine",
			from: enumType("ent/galaxy", "State", types.Typ[types.String],
				constDef{"StateActive", str("active")}),
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateActive", str("active")}, constDef{"GalaxyStatePending", str("pending")}),
			want: 0,
		},
		{
			name: "source value missing from target",
			from: enumType("ent/galaxy", "State", types.Typ[types.String],
				constDef{"StateActive", str("active")}, constDef{"StateArchived", str("archived")}),
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateActive", str("active")}),
			want: 1,
		},
		{
			name: "int enums with swapped values",
			from: enumType("ent/galaxy", "Code", types.Typ[types.Int],
				constDef{"CodeOK", num(1)}, constDef{"CodeError", num(2)}),
			to: enumType("store", "Code", types.Typ[types.Int],
				constDef{"CodeOK", num(2)}, constDef{"CodeError", num(1)}),
			want: 2,
		},
		{
			name: "int enums with matching values",
			from: enumType("ent/galaxy", "Code", types.Typ[types.Int],
				constDef{"CodeOK", num(1)}, constDef{"CodeError", num(2)}),
			to: enumType("store", "Code", types.Typ[types.Int],
				constDef{"CodeOK", num(1)}, constDef{"CodeError", num(2)}),
			want: 0,
		},
		{
			name: "differently named constants are compared by value only",
			from: enumType("ent/galaxy", "State", types.Typ[types.String],
				constDef{"StateActive", str("active")}),
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateEnabled", str("active")}),
			want: 0,
		},
		{
			name: "target without constants is not an enum",
			from: enumType("ent/galaxy", "State", types.Typ[types.String],
				constDef{"StateActive", str("active")}),
			to:   enumType("store", "GalaxyState", types.Typ[types.String]),
			want: 0,
		},
		{
			name: "source without constants is not an enum",
			from: enumType("ent/galaxy", "State", types.Typ[types.String]),
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateActive", str("active")}),
			want: 0,
		},
		{
			name: "unnamed types are not enums",
			from: nil,
			to: enumType("store", "GalaxyState", types.Typ[types.String],
				constDef{"GalaxyStateActive", str("active")}),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := tt.from
			if from == nil {
				from = types.Typ[types.String]
			}

			got := checkEnumConversion(from, tt.to)
			if len(got) != tt.want {
				t.Errorf("checkEnumConversion() returned %d problems, want %d: %v", len(got), tt.want, got)
			}
		})
	}
}

func TestNormalizeEnumName(t *testing.T) {
	tests := []struct {
		constName string
		typeName  string
		want      string
	}{
		{"StateActive", "State", "ACTIVE"},
		{"GalaxyStateActive", "GalaxyState", "ACTIVE"},
		{"Transition_State_ACTIVE", "Transition_State", "ACTIVE"},
		{"ACTIVE", "State", "ACTIVE"},
		{"state_active", "State", "STATEACTIVE"},
		// a constant named exactly like its type keeps its name
		{"State", "State", "STATE"},
		{"Active", "", "ACTIVE"},
	}

	for _, tt := range tests {
		t.Run(tt.constName, func(t *testing.T) {
			if got := normalizeEnumName(tt.constName, tt.typeName); got != tt.want {
				t.Errorf("normalizeEnumName(%q, %q) = %q, want %q", tt.constName, tt.typeName, got, tt.want)
			}
		})
	}
}
