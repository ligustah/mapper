package mapper

import "testing"

func TestMethodDirective(t *testing.T) {
	got := method(t, mustGenerate(t, "methods"), "ToResult")

	assertContains(t, got,
		"Passed: in.Passed()",
		"tmp2, err := tmp1.Decode()",
		"if err != nil { return nil, err }",
	)
}

// TestPointerResults covers a method or function that hands back a pointer already,
// on a pointer source: the pointer is kept, not addressed a second time.
func TestPointerResults(t *testing.T) {
	got := method(t, mustGenerate(t, "pointermethods"), "ToTarget")

	assertContains(t, got,
		"var tmp0 *pointermethods.Ref",
		"tmp2, err := tmp1.Encoded()",
		"tmp0 = tmp2",
		"tmp3 := make([]*pointermethods.Ref, len(in.Bodies))",
		"tmp7 = pointermethods.Wrap(tmp9)",
		"tmp10[k] = v.String()",
	)
}

// TestMethodResultConverts covers a method result of another type than the field,
// which goes through the conversion a field of that type would.
func TestMethodResultConverts(t *testing.T) {
	assertContains(t, method(t, mustGenerate(t, "methods"), "ToResult"),
		"StopsAt: int32(in.Ceiling())",
	)
}
