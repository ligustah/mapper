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
