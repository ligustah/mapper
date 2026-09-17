package mapper

import (
	"go/token"
	"go/types"
	"strconv"
	"testing"
)

func Test_splitPackageAndType(t *testing.T) {
	tests := []struct {
		name           string
		wantCanonical  string
		wantPkg        string
		wantTypeName   string
		wantMethodName string
		wantPointer    bool
		wantErr        bool
	}{
		{
			name:          "*google.golang.org/protobuf/types/known/timestamppb.Timestamp",
			wantCanonical: "*google.golang.org/protobuf/types/known/timestamppb.Timestamp",
			wantPkg:       "google.golang.org/protobuf/types/known/timestamppb",
			wantTypeName:  "Timestamp",
			wantPointer:   true,
		},
		{
			name:          "time.Time",
			wantCanonical: "time.Time",
			wantPkg:       "time",
			wantTypeName:  "Time",
		},
		{
			name:          "  *time.Time    ",
			wantCanonical: "*time.Time",
			wantPkg:       "time",
			wantTypeName:  "Time",
			wantPointer:   true,
		},
		{
			name:           "time.(*Time).String",
			wantCanonical:  "time.(*Time).String",
			wantPkg:        "time",
			wantTypeName:   "Time",
			wantMethodName: "String",
			wantPointer:    true,
		},
		{
			name:           "time.(Time).String",
			wantCanonical:  "time.(Time).String",
			wantPkg:        "time",
			wantTypeName:   "Time",
			wantMethodName: "String",
			wantPointer:    false,
		},
		{
			name:    "*time.Time.String",
			wantErr: true,
		},
		{
			name:    "time.Time.String.Foo",
			wantErr: true,
		},
		{
			name:    "*google.golang.org/protobuf/types/known/timestamppb",
			wantErr: true,
		},
		{
			name:    "Time",
			wantErr: true,
		}, {
			name:    "",
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotCanonical, gotPkg, gotTypeName, gotMethodName, gotPointer, err := splitPackageAndType(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitPackageAndType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// all other return values are undefined, when there is an error, so no need to check further
			if err != nil {
				return
			}

			if gotCanonical != tt.wantCanonical {
				t.Errorf("splitPackageAndType() gotCanonical = %v, want %v", gotCanonical, tt.wantCanonical)
			}
			if gotPkg != tt.wantPkg {
				t.Errorf("splitPackageAndType() gotPkg = %v, want %v", gotPkg, tt.wantPkg)
			}
			if gotTypeName != tt.wantTypeName {
				t.Errorf("splitPackageAndType() gotTypeName = %v, want %v", gotTypeName, tt.wantTypeName)
			}
			if gotMethodName != tt.wantMethodName {
				t.Errorf("splitPackageAndType() gotMethodName = %v, want %v", gotMethodName, tt.wantMethodName)
			}
			if gotPointer != tt.wantPointer {
				t.Errorf("splitPackageAndType() gotPointer = %v, want %v", gotPointer, tt.wantPointer)
			}
		})
	}
}

func TestQualifyReference(t *testing.T) {
	const pkg = "example.com/app/mapping"
	imports := map[string]string{
		"proto": "example.com/app/proto",
		"pb":    "google.golang.org/protobuf/types/known/timestamppb",
		"time":  "time",
	}

	tests := []struct {
		ref  string
		want string
	}{
		{"Ref", pkg + ".Ref"},
		{" Ref ", pkg + ".Ref"},
		{"(*Row).Name", pkg + ".(*Row).Name"},
		{"(Row).Name", pkg + ".(Row).Name"},
		{"proto.NewRef", "example.com/app/proto.NewRef"},
		{"proto.(*Ref).String", "example.com/app/proto.(*Ref).String"},
		{"pb.New", "google.golang.org/protobuf/types/known/timestamppb.New"},
		{"time.Unix", "time.Unix"},
		{"strconv.Itoa", "strconv.Itoa"},
		{"example.com/app/proto.NewRef", "example.com/app/proto.NewRef"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			if got := qualifyReference(tt.ref, pkg, imports); got != tt.want {
				t.Errorf("qualifyReference(%q) = %q, want %q", tt.ref, got, tt.want)
			}
		})
	}
}

// TestGoTypeRef guards that a type is spelled with everything wrapped around it.
// qual() names only the named type, so using it where a full type reference is
// needed silently drops the star off a `[]*T` element and generates `[]T`.
func TestGoTypeRef(t *testing.T) {
	storePkg := types.NewPackage("example.com/store", "store")
	point := named(storePkg, "Point", types.NewStruct(
		[]*types.Var{types.NewField(token.NoPos, nil, "Name", types.Typ[types.String], false)}, nil))
	items := named(storePkg, "Items", types.NewSlice(types.NewPointer(point)))

	tests := []struct {
		name string
		typ  types.Type
		want string
	}{
		{"named struct", point, "store.Point"},
		{"pointer to named struct", types.NewPointer(point), "*store.Point"},
		{"slice of named structs", types.NewSlice(point), "[]store.Point"},
		{"slice of pointers", types.NewSlice(types.NewPointer(point)), "[]*store.Point"},
		{"named slice of pointers", items, "store.Items"},
		{"basic", types.Typ[types.String], "string"},
		{"array", types.NewArray(types.Typ[types.Byte], 4), "[4]uint8"},
		{"map to pointers", types.NewMap(types.Typ[types.String], types.NewPointer(point)), "map[string]*store.Point"},
		{"pointer to pointer", types.NewPointer(types.NewPointer(point)), "**store.Point"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGoType(tt.typ).ref().GoString(); got != tt.want {
				t.Errorf("ref() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAssignableToAcrossLoads guards the `using:` resolution: the converter's
// package is loaded separately from the package being generated, so the same
// declaration comes back as two distinct types and go/types compares named types
// by identity.
func TestAssignableToAcrossLoads(t *testing.T) {
	first, err := LookupType("time.Time")
	if err != nil {
		t.Fatalf("LookupType() error = %v", err)
	}

	second, err := LookupType("time.Time")
	if err != nil {
		t.Fatalf("LookupType() error = %v", err)
	}

	if types.AssignableTo(first.Raw, second.Raw) {
		t.Skip("both loads returned the same type, there is nothing to guard here")
	}

	if !assignableTo(first.Raw, second.Raw) {
		t.Errorf("assignableTo(%v, %v) = false, want true", first, second)
	}

	other, err := LookupType("time.Duration")
	if err != nil {
		t.Fatalf("LookupType() error = %v", err)
	}

	if assignableTo(first.Raw, other.Raw) {
		t.Errorf("assignableTo(%v, %v) = true, want false", first, other)
	}
}
