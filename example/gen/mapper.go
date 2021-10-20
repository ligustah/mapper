package gen

import (
	"fmt"

	example "github.com/ligustah/mapper/example"
	models "github.com/ligustah/mapper/example/models"
	protos "github.com/ligustah/mapper/example/protos"
	null "github.com/volatiletech/null"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (m MapperImpl) MapJob(job *models.Job) *protos.Job {

	return &protos.Job{
		Id:         job.ID,
		CreatedAt:  timestamppb.New(job.CreatedAt),
		RunAt:      timestamppb.New(job.RunAt),
		Payload:    m.MapPayload(job),
		Queue:      job.Queue,
		Group:      job.Group.String,
		ExternalId: job.ExternalID.String,
	}
}

func (m MapperImpl) MapJobReverse(job *protos.Job) *models.Job {

	return &models.Job{
		ID:         job.Id,
		CreatedAt:  job.CreatedAt.AsTime(),
		RunAt:      job.RunAt.AsTime(),
		Group:      null.StringFrom(job.Group),
		Queue:      job.Queue,
		DataType:   job.Payload.ContentType,
		Data:       job.Payload.Body,
		ExternalID: null.StringFrom(job.ExternalId),
	}
}

func (m MapperImpl) MapPayload(job *models.Job) *protos.Data {

	return &protos.Data{
		ContentType: job.DataType,
		Body:        job.Data,
	}
}

func (m MapperImpl) MapTransition(transition *models.Transition) (*protos.Transition, error) {
	var tmp0 protos.Transition_State
	if val, ok := protos.Transition_State_value[transition.State]; ok {
		tmp0 = protos.Transition_State(val)
	} else {
		return nil, fmt.Errorf("invalid integer value for enum Transition_State: %d", transition.State)
	}

	return &protos.Transition{
		State:    tmp0,
		Time:     timestamppb.New(transition.CreatedAt),
		RetryAt:  timestamppb.New(transition.RetryAt.Time),
		Attempts: int32(transition.Attempts),
	}, nil
}

func (m MapperImpl) TransitionProtoToModel(transition *protos.Transition) (*models.Transition, error) {
	var tmp0 string
	if str, ok := protos.Transition_State_name[int32(transition.State)]; ok {
		tmp0 = str
	} else {
		return nil, fmt.Errorf("invalid string value for enum Transition_State: %d", transition.State)
	}

	return &models.Transition{
		CreatedAt: transition.Time.AsTime(),
		State:     tmp0,
		Attempts:  int(transition.Attempts),
		RetryAt:   null.TimeFrom(transition.RetryAt.AsTime()),
	}, nil
}

type MapperImpl struct{}

var _ example.Mapper = (*MapperImpl)(nil)
