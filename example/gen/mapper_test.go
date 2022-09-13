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
					t.Errorf("MapJob() returned unexpected result")
					t.Errorf(diff)
				}
			})

			t.Run("MapJobReverse", func(t *testing.T) {
				got := m.MapJobReverse(tt.pJob)
				if diff := cmp.Diff(got, tt.mJob, diffOpts); diff != "" {
					t.Errorf("MapJobReverse() returned unexpected result")
					t.Errorf(diff)
				}
			})
		})
	}
}
