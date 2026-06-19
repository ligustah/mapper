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

// AdvancedMapper demonstrates all supported features
// mapper:generate
type AdvancedMapper interface {
	// Basic field mapping with type conversion (int -> int32, time.Time -> string)
	// field:CreatedAt using:"github.com/ligustah/mapper/example.FormatTime"
	MapUser(user *models.User) *protos.UserProto

	// Reverse mapping with using: directive for string -> time.Time
	// field:CreatedAt using:"github.com/ligustah/mapper/example.ParseTime"
	MapUserReverse(user *protos.UserProto) (*models.User, error)

	// Slice and map field mapping
	// Scores: []int -> []int32 (automatic element conversion)
	// Metadata: map[string]int -> map[string]int64 (automatic element conversion)
	MapUserWithRoles(user *models.UserWithRoles) *protos.UserWithRolesProto

	// Reverse mapping for slices and maps
	// Scores: []int32 -> []int (automatic element conversion)
	// Metadata: map[string]int64 -> map[string]int (automatic element conversion)
	MapUserWithRolesReverse(user *protos.UserWithRolesProto) *models.UserWithRoles

	// Array field mapping ([4]byte -> [4]byte)
	MapArrayExample(src *models.ArrayExample) *protos.ArrayExampleProto

	// Reverse array mapping
	MapArrayExampleReverse(src *protos.ArrayExampleProto) *models.ArrayExample

	// using: directive for custom conversion (int64 -> string)
	// field:Timestamp using:"github.com/ligustah/mapper/example.ConvertTimestamp"
	MapCustomConvert(src *models.CustomConvertExample) *protos.CustomConvertExampleProto

	// Reverse using: directive with error handling (string -> int64)
	// field:Timestamp using:"github.com/ligustah/mapper/example.ParseTimestamp"
	MapCustomConvertReverse(src *protos.CustomConvertExampleProto) (*models.CustomConvertExample, error)

	// Named slice type mapping (StringSlice -> []string)
	MapNamedSlice(src *models.NamedSliceExample) *protos.NamedSliceExampleProto

	// Reverse named slice mapping ([]string -> StringSlice)
	MapNamedSliceReverse(src *protos.NamedSliceExampleProto) *models.NamedSliceExample
}
