package derive

type Goal struct {
	Kind  string
	Count int
}

type Node struct {
	Name     string
	Children []*Node
}

type Plan struct {
	ID       string
	Goals    []Goal
	Primary  Goal
	Fallback *Goal
	Root     *Node
	Host     string
	Port     int
	Size     string
}

type GoalProto struct {
	Kind  string
	Count int32
}

type NodeProto struct {
	Name     string
	Children []*NodeProto
}

type HostKey struct {
	Host string
	Port int32
}

type Sized struct {
	Size int
}

type PlanProto struct {
	ID       string
	Goals    []*GoalProto
	Primary  GoalProto
	Fallback *GoalProto
	Root     *NodeProto
	Key      HostKey
	Sized    Sized
}

// mapper:generate
type Mapper interface {
	// field:Key from:"."
	// field:Sized from:"."
	ToProto(in *Plan) (*PlanProto, error)
}

type Wrapper struct {
	Goals []Goal
}

type Holder struct {
	Goals []GoalProto
}

// mapper:generate
type Other interface {
	ToHolder(in *Wrapper) *Holder
}
