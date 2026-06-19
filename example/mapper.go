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
	// field:CreatedAt using:"github.com/ligustah/mapper/example.FormatTime"
	MapUser(user *models.User) *protos.UserProto

	// field:CreatedAt using:"github.com/ligustah/mapper/example.ParseTime"
	MapUserReverse(user *protos.UserProto) (*models.User, error)

	MapUserWithRoles(user *models.UserWithRoles) *protos.UserWithRolesProto

	MapUserWithRolesReverse(user *protos.UserWithRolesProto) *models.UserWithRoles

	MapArrayExample(src *models.ArrayExample) *protos.ArrayExampleProto

	MapArrayExampleReverse(src *protos.ArrayExampleProto) *models.ArrayExample

	// field:Timestamp using:"github.com/ligustah/mapper/example.ConvertTimestamp"
	MapCustomConvert(src *models.CustomConvertExample) *protos.CustomConvertExampleProto

	// field:Timestamp using:"github.com/ligustah/mapper/example.ParseTimestamp"
	MapCustomConvertReverse(src *protos.CustomConvertExampleProto) (*models.CustomConvertExample, error)

	MapNamedSlice(src *models.NamedSliceExample) *protos.NamedSliceExampleProto

	MapNamedSliceReverse(src *protos.NamedSliceExampleProto) *models.NamedSliceExample
}
