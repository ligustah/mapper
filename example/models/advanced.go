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
