package models

import "time"

// User demonstrates basic struct mapping
type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
}

// UserWithRoles demonstrates slice and map fields
type UserWithRoles struct {
	ID       string
	Name     string
	Roles    []string
	Tags     map[string]string
	Scores   []int
	Metadata map[string]int
}

// ArrayExample demonstrates fixed-size array support
type ArrayExample struct {
	ID   string
	Data [4]byte
	Name string
}

// ArrayConvertExample demonstrates a fixed-size array whose elements need a
// conversion, rather than being assignable as they are
type ArrayConvertExample struct {
	ID     string
	Values [3]int
}

// CustomConvertExample demonstrates the using: directive
type CustomConvertExample struct {
	ID        string
	Value     string
	Timestamp int64
}

// NamedSliceExample demonstrates named slice type support
type NamedSliceExample struct {
	ID    string
	Items StringSlice
}

// StringSlice is a named slice type
type StringSlice []string

// Status is a named string enum. Its values line up with protos.StatusProto, so a
// conversion between the two is derived automatically without any annotation.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// Visibility is a named string enum whose counterpart uses different values, so a
// mapping has to be spelled out with the map: directive
type Visibility string

const (
	VisibilityUnspecified Visibility = ""
	VisibilityPublic      Visibility = "public"
	VisibilityPrivate     Visibility = "private"
)

// EnumExample demonstrates enum mapping
type EnumExample struct {
	ID         string
	Status     Status
	Visibility Visibility
}

// Item is the element of a slice of pointers
type Item struct {
	Name  string
	Count int
}

// PointerSliceExample demonstrates a repeated field of pointers, the shape a
// protobuf `repeated Item` message field takes on the Go side
type PointerSliceExample struct {
	ID    string
	Items []*Item
}

// Snapshot is a named struct handed to a converter through the using: directive
type Snapshot struct {
	ID   string
	Tick int64
}

// SnapshotExample demonstrates using: with a converter over a named struct
type SnapshotExample struct {
	ID       string
	Snapshot *Snapshot
}

// Detail is an optional nested value, mapped through a value method
type Detail struct {
	Note   string
	Weight int
}

// OptionalExample demonstrates optional nested values: a nil field stays nil
type OptionalExample struct {
	ID     string
	Detail *Detail
	Item   *Item
}

// Event demonstrates conversions the generator derives on its own: a time read
// through a method, a string carried as bytes and an int onto a named integer type
type Event struct {
	ID    string
	At    time.Time
	Owner string
	Level int
}
