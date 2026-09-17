package example

//go:generate go run ../main.go .

import (
	"github.com/ligustah/mapper/example/models"
	"github.com/ligustah/mapper/example/protos"
)

// mapper:generate
type Mapper interface {
	// field:Payload from:"."
	// field:Transitions from:"-"
	MapJob(job *models.Job) *protos.Job

	// field:Time from:"CreatedAt"
	MapTransition(transition *models.Transition) (*protos.Transition, error)

	// field:ID from:"-"
	// field:JobID from:"-"
	// field:R from:"-"
	// field:L from:"-"
	// field:CreatedAt from:"Time"
	TransitionProtoToModel(transition *protos.Transition) (*models.Transition, error)

	// field:ContentType from:"DataType"
	// field:Body from:"Data"
	MapPayload(job *models.Job) *protos.Data

	// field:DataType from:"Payload.ContentType"
	// field:Data from:"Payload.Body"
	// field:UpdatedAt from:"-"
	// field:CronID from:"-"
	// field:R from:"-"
	// field:L from:"-"
	MapJobReverse(job *protos.Job) *models.Job
}

// AdvancedMapper demonstrates all supported mapper features:
// - Basic field mapping with automatic type conversion (int -> int32)
// - Custom converters using the `using:` directive
// - Slice and map field mapping with automatic element conversion
// - Fixed-size array support ([4]byte -> [4]byte)
// - Named slice type support (StringSlice -> []string)
// - Error handling for converters that can fail
// mapper:generate
type AdvancedMapper interface {
	// A bare name in using: refers to the package declaring the interface, and
	// `alias.Name` to one of the file's imports. The full import path still works.
	// field:CreatedAt using:"FormatTime"
	MapUser(user *models.User) *protos.UserProto

	// field:CreatedAt using:"github.com/ligustah/mapper/example.ParseTime"
	MapUserReverse(user *protos.UserProto) (*models.User, error)

	MapUserWithRoles(user *models.UserWithRoles) *protos.UserWithRolesProto

	MapUserWithRolesReverse(user *protos.UserWithRolesProto) *models.UserWithRoles

	MapArrayExample(src *models.ArrayExample) *protos.ArrayExampleProto

	MapArrayExampleReverse(src *protos.ArrayExampleProto) *models.ArrayExample

	// the elements are not assignable as they are, so every one of them is converted
	MapArrayConvert(src *models.ArrayConvertExample) *protos.ArrayConvertExampleProto

	// field:Timestamp using:"github.com/ligustah/mapper/example.ConvertTimestamp"
	MapCustomConvert(src *models.CustomConvertExample) *protos.CustomConvertExampleProto

	// field:Timestamp using:"github.com/ligustah/mapper/example.ParseTimestamp"
	MapCustomConvertReverse(src *protos.CustomConvertExampleProto) (*models.CustomConvertExample, error)

	MapNamedSlice(src *models.NamedSliceExample) *protos.NamedSliceExampleProto

	MapNamedSliceReverse(src *protos.NamedSliceExampleProto) *models.NamedSliceExample

	// Status needs no annotation, both sides are named string types declaring the
	// same values, so the conversion is derived. Visibility is numbered on the proto
	// side, so the values are listed explicitly. Either side of a pair can be spelled
	// as the name of a constant or as its value.
	// field:Visibility map:"VisibilityUnspecified=VisibilityProtoUnspecified,public=VisibilityProtoPublic,private=VisibilityProtoPrivate"
	MapEnumExample(src *models.EnumExample) (*protos.EnumExampleProto, error)

	// field:Visibility map:"VisibilityProtoUnspecified=VisibilityUnspecified,VisibilityProtoPublic=public,VisibilityProtoPrivate=private"
	MapEnumExampleReverse(src *protos.EnumExampleProto) (*models.EnumExample, error)

	MapItem(item *models.Item) *protos.ItemProto

	// Items is []*models.Item on one side and []*protos.ItemProto on the other, the
	// shape a repeated message field takes, so the elements stay pointers
	MapPointerSlice(src *models.PointerSliceExample) *protos.PointerSliceExampleProto

	// Both converters take a named struct this package already names, from a package
	// the generator loads on its own: one off a field, one off the whole input
	// field:Summary from:"Snapshot" using:"github.com/ligustah/mapper/example/converters.SummarizeSnapshot"
	// field:Digest from:"." using:"github.com/ligustah/mapper/example/converters.DigestExample"
	MapSnapshotExample(src *models.SnapshotExample) *protos.SnapshotExampleProto

	MapDetail(detail models.Detail) protos.DetailProto

	// Detail is *models.Detail on one side and *protos.DetailProto on the other, but
	// only a value method exists for the pair, so the generator wraps it in a nil
	// check and a nil field stays nil. Item has a pointer method, that one is called
	// as it is and guards nil itself.
	MapOptionalExample(src *models.OptionalExample) *protos.OptionalExampleProto

	// AtNs is read off a method of the source value. Owner and Level need no
	// annotation: string -> []byte and conversions between integer types, named or
	// not, are derived. Neither does OwnerId, a field name that has no exact match
	// falls back to a case-insensitive one, so it finds OwnerID.
	// field:AtNs from:"At" method:"UnixNano"
	MapEvent(src *models.Event) *protos.EventProto
}
