package protos

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
