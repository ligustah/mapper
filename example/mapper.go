package example

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
