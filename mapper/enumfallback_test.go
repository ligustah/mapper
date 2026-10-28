package mapper

import (
	"strings"
	"testing"
)

func TestEnumMapFallback(t *testing.T) {
	got := method(t, mustGenerate(t, "enumfallback"), "ToMessage")

	assertContains(t, got,
		"func (m MapperImpl) ToMessage(in *enumfallback.Row) *enumfallback.Message {",
		// only a fallback: the rest is paired by name, past the shared SET_ prefix
		"case enumfallback.StateActive: tmp0 = enumfallback.SetState_SET_ACTIVE",
		"case enumfallback.StateClosing: tmp0 = enumfallback.SetState_SET_CLOSING",
		"default: tmp0 = enumfallback.SetState_SET_UNSPECIFIED",
		`case enumfallback.KindFile: tmp1 = "file"`,
		`default: tmp1 = "unknown"`,
		"case true: tmp2 = enumfallback.KindDir",
		"default: tmp2 = enumfallback.KindFile",
	)

	if strings.Contains(got, "Errorf") {
		t.Errorf("a fallback leaves no error branch:\n%s", got)
	}
}
