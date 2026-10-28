package mapper

import (
	"os/exec"
	"testing"
)

func TestDeriveStruct(t *testing.T) {
	out := mustGenerate(t, "derive")

	assertContains(t, method(t, out, "mapGoalToGoalProto"),
		"func (m MapperImpl) mapGoalToGoalProto(in derive.Goal) derive.GoalProto {",
		"Count: int32(in.Count)",
	)
	assertContains(t, method(t, out, "mapNodeToNodeProto"),
		"func (m MapperImpl) mapNodeToNodeProto(in derive.Node) derive.NodeProto {",
	)
	assertContains(t, method(t, out, "mapPlanToSized"),
		"func (m MapperImpl) mapPlanToSized(in derive.Plan) (derive.Sized, error) {",
	)
	assertContains(t, method(t, out, "mapPlanToHostKey"),
		"Host: in.Host",
		"Port: int32(in.Port)",
	)

	// every interface derives its own, a helper on another implementation is
	// out of reach
	assertContains(t, method(t, out, "ToHolder"), "m.mapGoalToGoalProto(in.Goals[i])")
	assertContains(t, out, "func (m OtherImpl) mapGoalToGoalProto(in derive.Goal) derive.GoalProto {")
}

// TestDeriveNoCopy covers structs go vet won't let be copied, protobuf messages:
// their helpers take and return pointers, and nothing dereferences one.
func TestDeriveNoCopy(t *testing.T) {
	out := mustGenerate(t, "nocopy")

	assertContains(t, out,
		"func (m MapperImpl) mapBracketToStoreBracket(in *nocopy.Bracket) *nocopy.StoreBracket {",
		"func (m MapperImpl) mapStoreBracketToBracket(in *nocopy.StoreBracket) *nocopy.Bracket {",
	)

	vet := exec.Command("go", "vet", "./testdata/nocopy/gen")
	if output, err := vet.CombinedOutput(); err != nil {
		t.Errorf("go vet on the generated code: %v\n%s\n%s", err, output, out)
	}
}
