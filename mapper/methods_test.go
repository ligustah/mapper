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

// TestConversionMethod covers a method whose result is no struct: its input is
// converted as a whole, through the interface's other methods but never itself.
func TestConversionMethod(t *testing.T) {
	out := mustGenerate(t, "convmethod")

	views := method(t, out, "ToViews")
	assertContains(t, views,
		"func (m MapperImpl) ToViews(in []*convmethod.Item) []*convmethod.ItemView {",
		"tmp0[i] = m.ToView(in[i])",
		"return tmp0",
	)
	assertContains(t, method(t, out, "ToState"),
		"case convmethod.StateOpen: tmp0 = convmethod.StateView_OPEN",
		"return tmp0, nil",
	)
}
