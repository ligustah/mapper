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
