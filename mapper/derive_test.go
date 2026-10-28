package mapper

import "testing"

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
