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
