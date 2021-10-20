package mapper

import (
	"strconv"
	"testing"
)

func TestLookupType(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "time.Time"},
		{name: "*google.golang.org/protobuf/types/known/timestamppb.Timestamp"},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := LookupType(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Raw.String() != tt.name {
				t.Errorf("LookupType() got = %v, want %v", got.Raw.String(), tt.name)
			}
		})
	}
}
