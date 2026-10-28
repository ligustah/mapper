package mapper

import (
	"strings"
	"testing"
)

func TestSkipUnmatched(t *testing.T) {
	out := mustGenerate(t, "skipunmatched")

	for _, name := range []string{"ToRow", "ToRowSparse"} {
		got := method(t, out, name)
		assertContains(t, got, "ID: in.ID", "Name: in.Name")
		if strings.Contains(got, "Edges") || strings.Contains(got, "Extra") {
			t.Errorf("%s should leave unmatched fields out:\n%s", name, got)
		}
	}

	fill := method(t, out, "Fill")
	assertContains(t, fill, "b.SetID(in.ID)")
	if strings.Contains(fill, "SetOwner") {
		t.Errorf("Fill should leave the unmatched setter out:\n%s", fill)
	}
}

func TestSkipUnmatchedKeepsDirectivesStrict(t *testing.T) {
	_, err := generate(t, "skipunmatched_directive")
	if err == nil || !strings.Contains(err.Error(), "could not find source for output Name") {
		t.Fatalf("a directive naming a missing source should still fail, got %v", err)
	}
}

func TestInline(t *testing.T) {
	out := mustGenerate(t, "inline")

	// Report comes first, and is a pointer: its fields are read behind a nil check
	assertContains(t, method(t, out, "ToRevenue"),
		"if in.Report != nil { tmp0 = in.Report.Total }",
		"if in.Report != nil { tmp1 = in.Report.Count }",
		"ID: in.ID",
		"Name: in.Host.Name",
	)

	// on the interface, and a directive still wins over it
	assertContains(t, method(t, out, "ToHostRevenue"),
		"if in.Report != nil { tmp0 = in.Report.Count }",
		"Name: in.Host.Name",
	)
}

func TestFieldFromParameter(t *testing.T) {
	assertContains(t, method(t, mustGenerate(t, "params"), "ToRow"),
		"Name: in.Name",
		"ClosedAtNs: closedAtNs",
		"Owner: owner",
		"Actor: actor",
	)
}
