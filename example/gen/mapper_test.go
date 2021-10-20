package gen

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"

	"github.com/ligustah/mapper/example"
	"github.com/ligustah/mapper/example/models"
	"github.com/ligustah/mapper/example/protos"
	"github.com/volatiletech/null"
)

func TestMapperImpl_MapJob(t *testing.T) {
	createdAt := time.Now()
	updatedAt := createdAt.Add(1234 * time.Hour)
	runAt := createdAt.Add(4466 * time.Hour)

	tests := []struct {
		name string
		job  *models.Job
		want *protos.Job
	}{
		{
			name: "1",
			job: &models.Job{
				ID:         "1234",
				CreatedAt:  createdAt,
				UpdatedAt:  null.NewTime(updatedAt, true),
				RunAt:      runAt,
				Group:      null.String{},
				Queue:      "queue",
				DataType:   "type",
				Data:       []byte("data"),
				ExternalID: null.NewString("external", true),
				CronID:     null.String{},
			},
			want: &protos.Job{
				Id:        "1234",
				CreatedAt: timestamppb.New(createdAt),
				RunAt:     timestamppb.New(runAt),
				//Payload: &protos.Data{
				//	ContentType: "type",
				//	Body:        []byte("data"),
				//},
				Queue:       "queue",
				Transitions: nil,
				Group:       "",
				ExternalId:  "external",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MapperImpl{}
			got := m.MapJob(tt.job)
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(protos.Job{}, timestamppb.Timestamp{})); diff != "" {
				t.Errorf("MapJob() returned unexpected result")
				t.Errorf(diff)
			}
		})
	}
}

func TestMapperImpl_MapJobReverse(t *testing.T) {
	type args struct {
		job *protos.Job
	}
	tests := []struct {
		name string
		args args
		want *models.Job
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MapperImpl{}
			if got := m.MapJobReverse(tt.args.job); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapJobReverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func strPointer(input string) *string {
	return &input
}

func TestMapperImpl_MapNumericToString(t *testing.T) {
	type args struct {
		numeric *example.SimpleNumeric
	}
	tests := []struct {
		name string
		have *example.SimpleNumeric
		want *example.SimpleString
	}{
		{"1",
			&example.SimpleNumeric{
				Test:    45672,
				Slice:   []int{1, 2, 3, 45, 6, 7, 8},
				Pointer: "",
			},
			&example.SimpleString{
				Value:   "45672",
				Slice:   []string{"1", "2", "3", "45", "6", "7", "8"},
				Pointer: strPointer(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MapperImpl{}
			got := m.MapNumericToString(tt.have)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("MapNumericToString()")
				t.Errorf(diff)
			}
		})
	}
}

func TestMapperImpl_MapStringToNumeric(t *testing.T) {
	type args struct {
		str *example.SimpleString
	}
	tests := []struct {
		name    string
		args    args
		want    *example.SimpleNumeric
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MapperImpl{}
			got, err := m.MapStringToNumeric(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapStringToNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapStringToNumeric() got = %v, want %v", got, tt.want)
			}
		})
	}
}
