package mapper

import (
	"strconv"
	"testing"
)

// TestParseOrderIsStable guards the reproducibility of the generated output.
// Interfaces are discovered by walking a map, so without sorting them the order
// they come back in, and with it the layout of the generated file, changes between
// runs of the same unchanged input.
func TestParseOrderIsStable(t *testing.T) {
	parser := NewParser()

	names := func(interfaces []*GoInterface) []string {
		out := make([]string, 0, len(interfaces))
		for _, iface := range interfaces {
			out = append(out, iface.Pkg+"."+iface.Name)
		}
		return out
	}

	first, err := parser.Parse("../example", ".")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(first) < 2 {
		t.Fatalf("expected the example package to declare at least two interfaces, got %d", len(first))
	}

	second, err := parser.Parse("../example", ".")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	got, want := names(first), names(second)
	if len(got) != len(want) {
		t.Fatalf("Parse() returned %d interfaces, then %d", len(got), len(want))
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("Parse() is not stable, got %v, then %v", got, want)
		}
	}

	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Errorf("Parse() returned unsorted interfaces: %q before %q", got[i-1], got[i])
		}
	}
}

func TestLookupType(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "time.Time"},
		{name: "*google.golang.org/protobuf/types/known/timestamppb.Timestamp"},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := LookupType(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Raw.String() != tt.name {
				t.Errorf("LookupType() got = %v, want %v", got.Raw.String(), tt.name)
			}
		})
	}
}
