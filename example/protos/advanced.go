package protos

import "google.golang.org/protobuf/types/known/timestamppb"

// UserProto demonstrates basic struct mapping
type UserProto struct {
	ID        string
	Name      string
	Email     string
	Age       int32
	CreatedAt string
}

// UserWithRolesProto demonstrates slice and map fields
type UserWithRolesProto struct {
	ID       string
	Name     string
	Roles    []string
	Tags     map[string]string
	Scores   []int32
	Metadata map[string]int64
}

// ArrayExampleProto demonstrates fixed-size array support
type ArrayExampleProto struct {
	ID   string
	Data [4]byte
	Name string
}

// ArrayConvertExampleProto demonstrates a fixed-size array of converted elements
type ArrayConvertExampleProto struct {
	ID     string
	Values [3]int32
}

// CustomConvertExampleProto demonstrates the using: directive
type CustomConvertExampleProto struct {
	ID        string
	Value     string
	Timestamp string
}

// NamedSliceExampleProto demonstrates named slice type support
type NamedSliceExampleProto struct {
	ID    string
	Items []string
}

// StatusProto declares the same values as models.Status
type StatusProto string

const (
	StatusProtoActive   StatusProto = "active"
	StatusProtoInactive StatusProto = "inactive"
)

// VisibilityProto is numbered rather than named like models.Visibility
type VisibilityProto int32

const (
	VisibilityProtoUnspecified VisibilityProto = 0
	VisibilityProtoPublic      VisibilityProto = 1
	VisibilityProtoPrivate     VisibilityProto = 2
)

// EnumExampleProto demonstrates enum mapping
type EnumExampleProto struct {
	ID         string
	Status     StatusProto
	Visibility VisibilityProto
}

// ItemProto is the element of a repeated field
type ItemProto struct {
	Name  string
	Count int32
}

// PointerSliceExampleProto demonstrates a repeated message field
type PointerSliceExampleProto struct {
	ID    string
	Items []*ItemProto
}

// SnapshotExampleProto demonstrates using: with a converter over a named struct
type SnapshotExampleProto struct {
	ID      string
	Summary string
	Digest  string
}

// DetailProto is an optional nested message
type DetailProto struct {
	Note   string
	Weight int32
}

// OptionalExampleProto demonstrates optional nested messages
type OptionalExampleProto struct {
	ID     string
	Detail *DetailProto
	Item   *ItemProto
}

// TaskPhase is numbered and named the way protoc names an enum's values
type TaskPhase int32

const (
	TaskPhase_TASK_PHASE_UNSPECIFIED TaskPhase = 0
	TaskPhase_TASK_PHASE_RUNNING     TaskPhase = 1
	TaskPhase_TASK_PHASE_DONE        TaskPhase = 2
)

// TaskPhase_name and TaskPhase_value are the lookup tables protoc emits for an enum
var (
	TaskPhase_name = map[int32]string{
		0: "TASK_PHASE_UNSPECIFIED",
		1: "TASK_PHASE_RUNNING",
		2: "TASK_PHASE_DONE",
	}
	TaskPhase_value = map[string]int32{
		"TASK_PHASE_UNSPECIFIED": 0,
		"TASK_PHASE_RUNNING":     1,
		"TASK_PHASE_DONE":        2,
	}
)

// Enum is the method protoc generates, which marks the type as a protobuf enum
func (x TaskPhase) Enum() *TaskPhase {
	return &x
}

// TaskProto demonstrates a numbered enum mapped onto a string one
type TaskProto struct {
	ID    string
	Phase TaskPhase
}

// OrderProto flattens a field reached through an edge
type OrderProto struct {
	ID           string
	CustomerName string
}

// Level is a numbered enum narrower than the model's plain int
type Level uint8

const (
	LevelInfo Level = iota
	LevelWarn
)

// EventProto demonstrates derived conversions onto bytes and a named integer type
type EventProto struct {
	ID          string
	AtNs        int64
	Owner       []byte
	OwnerId     int64
	Level       Level
	RetiredAtNs int64
	Deadline    *timestamppb.Timestamp
	Note        string
}
