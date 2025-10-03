package mapper

import (
	"strconv"
	"testing"
)

func Test_splitPackageAndType(t *testing.T) {
	tests := []struct {
		name          string
		wantCanonical string
		wantPkg       string
		wantTypeName  string
		wantPointer   bool
		wantErr       bool
	}{
		{
			name:          "*google.golang.org/protobuf/types/known/timestamppb.Timestamp",
			wantCanonical: "*google.golang.org/protobuf/types/known/timestamppb.Timestamp",
			wantPkg:       "google.golang.org/protobuf/types/known/timestamppb",
			wantTypeName:  "Timestamp",
			wantPointer:   true,
		},
		{
			name:          "time.Time",
			wantCanonical: "time.Time",
			wantPkg:       "time",
			wantTypeName:  "Time",
		},
		{
			name:          "  *time.Time    ",
			wantCanonical: "*time.Time",
			wantPkg:       "time",
			wantTypeName:  "Time",
			wantPointer:   true,
		},
		{
			name:    "*google.golang.org/protobuf/types/known/timestamppb",
			wantErr: true,
		},
		{
			name:    "Time",
			wantErr: true,
		}, {
			name:    "",
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			gotCanonical, gotPkg, gotTypeName, _, gotPointer, err := splitPackageAndType(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitPackageAndType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// all other return values are undefined, when there is an error, so no need to check further
			if err != nil {
				return
			}

			if gotCanonical != tt.wantCanonical {
				t.Errorf("splitPackageAndType() gotCanonical = %v, want %v", gotCanonical, tt.wantCanonical)
			}
			if gotPkg != tt.wantPkg {
				t.Errorf("splitPackageAndType() gotPkg = %v, want %v", gotPkg, tt.wantPkg)
			}
			if gotTypeName != tt.wantTypeName {
				t.Errorf("splitPackageAndType() gotTypeName = %v, want %v", gotTypeName, tt.wantTypeName)
			}
			if gotPointer != tt.wantPointer {
				t.Errorf("splitPackageAndType() gotPointer = %v, want %v", gotPointer, tt.wantPointer)
			}
		})
	}
}
