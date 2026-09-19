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

func TestParseEnumMap(t *testing.T) {
	tests := []struct {
		name      string
		directive string
		want      []enumMapEntry
		wantErr   bool
	}{
		{
			name:      "single pair",
			directive: "active=ENABLED",
			want:      []enumMapEntry{{"active", "ENABLED"}},
		},
		{
			name:      "several pairs with spaces",
			directive: "active=ENABLED, retired = DISABLED ",
			want:      []enumMapEntry{{"active", "ENABLED"}, {"retired", "DISABLED"}},
		},
		{
			name:      "constant names",
			directive: "StateActive=GalaxyStateEnabled",
			want:      []enumMapEntry{{"StateActive", "GalaxyStateEnabled"}},
		},
		{
			name:      "trailing comma is ignored",
			directive: "active=ENABLED,",
			want:      []enumMapEntry{{"active", "ENABLED"}},
		},
		{name: "missing separator", directive: "active", wantErr: true},
		{name: "empty target", directive: "active=", wantErr: true},
		{name: "empty source", directive: "=ENABLED", wantErr: true},
		{name: "duplicate source", directive: "active=ENABLED,active=DISABLED", wantErr: true},
		{name: "empty directive", directive: "", wantErr: true},
		{name: "only separators", directive: ",,", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnumMap(tt.directive)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseEnumMap(%q) error = %v, wantErr %v", tt.directive, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("parseEnumMap(%q) = %v, want %v", tt.directive, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLookupEnumConstant(t *testing.T) {
	stringEnum := enumType("ent/galaxy", "State", types.Typ[types.String],
		constDef{"StateActive", str("active")}, constDef{"StateRetired", str("retired")})
	intEnum := enumType("ent/galaxy", "Code", types.Typ[types.Int],
		constDef{"CodeOK", num(1)})
	plain := enumType("ent/galaxy", "Plain", types.Typ[types.String])

	tests := []struct {
		name       string
		typ        types.Type
		ref        string
		wantName   string
		wantFound  bool
		wantIsEnum bool
	}{
		{"by constant name", stringEnum, "StateActive", "StateActive", true, true},
		{"by value", stringEnum, "active", "StateActive", true, true},
		{"by value, second constant", stringEnum, "retired", "StateRetired", true, true},
		{"unknown reference", stringEnum, "nope", "", false, true},
		{"value of a different constant kind", stringEnum, "1", "", false, true},
		{"int enum by name", intEnum, "CodeOK", "CodeOK", true, true},
		{"int enum by value", intEnum, "1", "CodeOK", true, true},
		{"type without constants", plain, "anything", "", false, false},
		{"unnamed type", types.Typ[types.String], "anything", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found, isEnum := lookupEnumConstant(tt.typ, tt.ref)
			if found != tt.wantFound || isEnum != tt.wantIsEnum {
				t.Fatalf("lookupEnumConstant(%q) found = %v, isEnum = %v, want %v, %v",
					tt.ref, found, isEnum, tt.wantFound, tt.wantIsEnum)
			}
			if got.name != tt.wantName {
				t.Errorf("lookupEnumConstant(%q) = %q, want %q", tt.ref, got.name, tt.wantName)
			}
		})
	}
}

func TestEnumKey(t *testing.T) {
	tests := []struct {
		constName string
		typeName  string
		want      string
	}{
		{"StateActive", "State", "ACTIVE"},
		{"SetState_SET_STATE_ACTIVE", "SetState", "ACTIVE"},
		{"SetState_SET_STATE_UNSPECIFIED", "SetState", "UNSPECIFIED"},
		{"Commit_ACTIVE", "Commit_State", "ACTIVE"},
		{"Commit_STATE_ACTIVE", "Commit_State", "ACTIVE"},
		{"VisibilityProtoPublic", "VisibilityProto", "PUBLIC"},
		// a constant named exactly like its type keeps its name
		{"State", "State", "STATE"},
	}

	for _, tt := range tests {
		t.Run(tt.constName, func(t *testing.T) {
			if got := enumKey(tt.constName, tt.typeName); got != tt.want {
				t.Errorf("enumKey(%q, %q) = %q, want %q", tt.constName, tt.typeName, got, tt.want)
			}
		})
	}
}

// TestMatchEnumConstants covers pairing an ent style string enum with a protoc
// style numbered one: by name, by string value, and nothing for a constant without
// a unique counterpart.
func TestMatchEnumConstants(t *testing.T) {
	state := enumType("example.com/ent/set", "State", types.Typ[types.String],
		constDef{"StateActive", str("active")},
		constDef{"StateClosing", str("closing")},
		// matched by its value alone
		constDef{"StateGone", str("deleted")},
		constDef{"StateLost", str("lost")},
	)
	pbState := enumType("example.com/pb", "SetState", types.Typ[types.Int32],
		constDef{"SetState_SET_STATE_UNSPECIFIED", num(0)},
		constDef{"SetState_SET_STATE_ACTIVE", num(1)},
		constDef{"SetState_SET_STATE_CLOSING", num(2)},
		constDef{"SetState_SET_STATE_DELETED", num(3)},
	)

	pairs, unmatched := matchEnumConstants(state, pbState)

	got := map[string]string{}
	for _, p := range pairs {
		got[p.from.name] = p.to.name
	}

	want := map[string]string{
		"StateActive":  "SetState_SET_STATE_ACTIVE",
		"StateClosing": "SetState_SET_STATE_CLOSING",
		"StateGone":    "SetState_SET_STATE_DELETED",
	}
	if len(got) != len(want) {
		t.Fatalf("matchEnumConstants() pairs = %v, want %v", got, want)
	}
	for from, to := range want {
		if got[from] != to {
			t.Errorf("%s maps onto %q, want %q", from, got[from], to)
		}
	}

	if len(unmatched) != 1 || unmatched[0] != "StateLost" {
		t.Errorf("unmatched = %v, want [StateLost]", unmatched)
	}

	if pairs, _ := matchEnumConstants(state, types.Typ[types.Int32]); pairs != nil {
		t.Errorf("a type without constants should match nothing, got %v", pairs)
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
