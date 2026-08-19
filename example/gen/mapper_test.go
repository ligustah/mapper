package gen

import (
	"testing"
	"time"

	"github.com/ligustah/mapper/example/models"
	"github.com/ligustah/mapper/example/protos"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/volatiletech/null"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapperImpl(t *testing.T) {
	createdAt := time.Now()
	runAt := createdAt.Add(4466 * time.Hour)

	tests := []struct {
		name string
		mJob *models.Job
		pJob *protos.Job
	}{
		{
			name: "1",
			mJob: &models.Job{
				ID:         "1234",
				CreatedAt:  createdAt,
				RunAt:      runAt,
				Group:      null.String{},
				Queue:      "queue",
				DataType:   "type",
				Data:       []byte("data"),
				ExternalID: null.NewString("external", true),
				CronID:     null.String{},
			},
			pJob: &protos.Job{
				Id:        "1234",
				CreatedAt: timestamppb.New(createdAt),
				RunAt:     timestamppb.New(runAt),
				Payload: &protos.Data{
					ContentType: "type",
					Body:        []byte("data"),
				},
				Queue:       "queue",
				Transitions: nil,
				Group:       "",
				ExternalId:  "external",
			},
		},
	}

	m := MapperImpl{}
	diffOpts := cmpopts.IgnoreUnexported(protos.Job{}, protos.Data{}, timestamppb.Timestamp{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("MapJob", func(t *testing.T) {
				got := m.MapJob(tt.mJob)
				if diff := cmp.Diff(got, tt.pJob, diffOpts); diff != "" {
					t.Errorf("MapJob() returned unexpected result:\n%s", diff)
				}
			})

			t.Run("MapJobReverse", func(t *testing.T) {
				got := m.MapJobReverse(tt.pJob)
				if diff := cmp.Diff(got, tt.mJob, diffOpts); diff != "" {
					t.Errorf("MapJobReverse() returned unexpected result:\n%s", diff)
				}
			})
		})
	}
}

func TestMapTransition(t *testing.T) {
	createdAt := time.Now()
	retryAt := createdAt.Add(1 * time.Hour)

	m := MapperImpl{}

	t.Run("model to proto", func(t *testing.T) {
		mTransition := &models.Transition{
			ID:        1,
			JobID:     "job-123",
			CreatedAt: createdAt,
			State:     "queued",
			Attempts:  3,
			RetryAt:   null.NewTime(retryAt, true),
		}

		got, err := m.MapTransition(mTransition)
		if err != nil {
			t.Fatalf("MapTransition() error = %v", err)
		}

		if got.State != protos.Transition_queued {
			t.Errorf("MapTransition() State = %v, want %v", got.State, protos.Transition_queued)
		}
		if got.Attempts != 3 {
			t.Errorf("MapTransition() Attempts = %v, want 3", got.Attempts)
		}
		if got.Time.AsTime().Unix() != createdAt.Unix() {
			t.Errorf("MapTransition() Time = %v, want %v", got.Time.AsTime(), createdAt)
		}
	})

	t.Run("proto to model", func(t *testing.T) {
		pTransition := &protos.Transition{
			State:    protos.Transition_executing,
			Time:     timestamppb.New(createdAt),
			RetryAt:  timestamppb.New(retryAt),
			Attempts: 5,
		}

		got, err := m.TransitionProtoToModel(pTransition)
		if err != nil {
			t.Fatalf("TransitionProtoToModel() error = %v", err)
		}

		if got.State != "executing" {
			t.Errorf("TransitionProtoToModel() State = %v, want executing", got.State)
		}
		if got.Attempts != 5 {
			t.Errorf("TransitionProtoToModel() Attempts = %v, want 5", got.Attempts)
		}
		if got.CreatedAt.Unix() != createdAt.Unix() {
			t.Errorf("TransitionProtoToModel() CreatedAt = %v, want %v", got.CreatedAt, createdAt)
		}
	})

	t.Run("invalid enum string", func(t *testing.T) {
		mTransition := &models.Transition{
			State: "invalid_state",
		}

		_, err := m.MapTransition(mTransition)
		if err == nil {
			t.Error("MapTransition() expected error for invalid enum, got nil")
		}
	})
}

func TestMapPayload(t *testing.T) {
	m := MapperImpl{}

	mJob := &models.Job{
		DataType: "application/json",
		Data:     []byte(`{"key":"value"}`),
	}

	got := m.MapPayload(mJob)

	if got.ContentType != "application/json" {
		t.Errorf("MapPayload() ContentType = %v, want application/json", got.ContentType)
	}
	if string(got.Body) != `{"key":"value"}` {
		t.Errorf("MapPayload() Body = %v, want {\"key\":\"value\"}", string(got.Body))
	}
}
