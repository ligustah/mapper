package gen

import (
	"fmt"
	example "github.com/ligustah/mapper/example"
	models "github.com/ligustah/mapper/example/models"
	protos "github.com/ligustah/mapper/example/protos"
	null "github.com/volatiletech/null"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (m AdvancedMapperImpl) MapArrayExample(src *models.ArrayExample) *protos.ArrayExampleProto {

	return &protos.ArrayExampleProto{
		ID:   src.ID,
		Data: src.Data,
		Name: src.Name,
	}
}

func (m AdvancedMapperImpl) MapArrayExampleReverse(src *protos.ArrayExampleProto) *models.ArrayExample {

	return &models.ArrayExample{
		ID:   src.ID,
		Data: src.Data,
		Name: src.Name,
	}
}

func (m AdvancedMapperImpl) MapCustomConvert(src *models.CustomConvertExample) *protos.CustomConvertExampleProto {

	return &protos.CustomConvertExampleProto{
		ID:        src.ID,
		Value:     src.Value,
		Timestamp: example.ConvertTimestamp(src.Timestamp),
	}
}

func (m AdvancedMapperImpl) MapCustomConvertReverse(src *protos.CustomConvertExampleProto) (*models.CustomConvertExample, error) {
	tmp0, err := example.ParseTimestamp(src.Timestamp)
	if err != nil {
		return nil, err
	}

	return &models.CustomConvertExample{
		ID:        src.ID,
		Value:     src.Value,
		Timestamp: tmp0,
	}, nil
}

func (m AdvancedMapperImpl) MapNamedSlice(src *models.NamedSliceExample) *protos.NamedSliceExampleProto {

	return &protos.NamedSliceExampleProto{
		ID:    src.ID,
		Items: src.Items,
	}
}

func (m AdvancedMapperImpl) MapNamedSliceReverse(src *protos.NamedSliceExampleProto) *models.NamedSliceExample {

	return &models.NamedSliceExample{
		ID:    src.ID,
		Items: src.Items,
	}
}

func (m AdvancedMapperImpl) MapUser(user *models.User) *protos.UserProto {

	return &protos.UserProto{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Age:       int32(user.Age),
		CreatedAt: example.FormatTime(user.CreatedAt),
	}
}

func (m AdvancedMapperImpl) MapUserReverse(user *protos.UserProto) (*models.User, error) {
	tmp0, err := example.ParseTime(user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Age:       int(user.Age),
		CreatedAt: tmp0,
	}, nil
}

func (m AdvancedMapperImpl) MapUserWithRoles(user *models.UserWithRoles) *protos.UserWithRolesProto {
	tmp0 := make([]int32, len(user.Scores))
	for i := range user.Scores {
		tmp0[i] = int32(user.Scores[i])
	}

	tmp1 := make(map[string]int64)
	for k, v := range user.Metadata {
		tmp1[k] = int64(v)
	}

	return &protos.UserWithRolesProto{
		ID:       user.ID,
		Name:     user.Name,
		Roles:    user.Roles,
		Tags:     user.Tags,
		Scores:   tmp0,
		Metadata: tmp1,
	}
}

func (m AdvancedMapperImpl) MapUserWithRolesReverse(user *protos.UserWithRolesProto) *models.UserWithRoles {
	tmp0 := make([]int, len(user.Scores))
	for i := range user.Scores {
		tmp0[i] = int(user.Scores[i])
	}

	tmp1 := make(map[string]int)
	for k, v := range user.Metadata {
		tmp1[k] = int(v)
	}

	return &models.UserWithRoles{
		ID:       user.ID,
		Name:     user.Name,
		Roles:    user.Roles,
		Tags:     user.Tags,
		Scores:   tmp0,
		Metadata: tmp1,
	}
}

type AdvancedMapperImpl struct{}

var _ example.AdvancedMapper = (*AdvancedMapperImpl)(nil)

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
	tmp0 := job.Group
	tmp1 := job.ExternalId

	return &models.Job{
		ID:         job.Id,
		CreatedAt:  job.CreatedAt.AsTime(),
		RunAt:      job.RunAt.AsTime(),
		Group:      null.NewString(tmp0, !(tmp0 == "")),
		Queue:      job.Queue,
		DataType:   job.Payload.ContentType,
		Data:       job.Payload.Body,
		ExternalID: null.NewString(tmp1, !(tmp1 == "")),
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
		return nil, fmt.Errorf("invalid string value for enum Transition_State: %s", transition.State)
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
		return nil, fmt.Errorf("invalid integer value for enum Transition_State: %d", transition.State)
	}

	tmp1 := transition.RetryAt.AsTime()

	return &models.Transition{
		CreatedAt: transition.Time.AsTime(),
		State:     tmp0,
		Attempts:  int(transition.Attempts),
		RetryAt:   null.NewTime(tmp1, !(tmp1.IsZero())),
	}, nil
}

type MapperImpl struct{}

var _ example.Mapper = (*MapperImpl)(nil)
