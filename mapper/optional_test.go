package mapper

import (
	"strings"
	"testing"
)

func TestDefaultAndZeroNil(t *testing.T) {
	out := mustGenerate(t, "optional")

	assertContains(t, method(t, out, "ToRow"),
		"var tmp0 bool = true",
		"var tmp3 bool = false",
		`var tmp6 string = "anonymous"`,
		"var tmp9 optional.Kind = optional.KindDir",
		"var tmp12 int32 = -1",
		"tmp14 := *tmp13",
		"tmp12 = int32(tmp14)",
		"var tmp15 int = 100",
		"tmp18, err := optional.Limit(tmp17)",
		"var tmp19 int64 = 0",
		"tmp19 = tmp20.Seconds()",
		"var tmp21 *time.Time",
		"if !in.At.IsZero() {",
		"var tmp23 *time.Time",
		"if in.Created != 0 {",
		"tmp24 := optional.Unix(in.Created)",
		"tmp23 = &tmp24",
		"if in.Max != 0 {",
		"if in.Origin != (optional.Point{}) {",
	)

	// a default: keeps the plain setter, a zero:"nil" hands the nillable one its nil
	assertContains(t, method(t, out, "Fill"),
		"b.SetNillableMax(tmp0)",
		"var tmp2 bool = true",
		"b.SetOn(tmp2)",
	)
}

func TestDefaultAndZeroNilMisuse(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"optional_bad_default", "default: the source is a int, which is never nil"},
		{"optional_bad_zero", `zero:"nil" is for a pointer field`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generate(t, tt.name)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want an error containing %q", err, tt.want)
			}
		})
	}
}
