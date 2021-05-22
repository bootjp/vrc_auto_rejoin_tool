package vrcarjt

import (
	"reflect"
	"testing"
)

func Test_getVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    *Version
		wantErr bool
	}{
		{
			name: "Should return a valid Version for v2.7.0",
			args: args{
				"v2.7.0",
			},
			want: &Version{
				Major: 2,
				Minor: 7,
				Patch: 0,
			},
			wantErr: false,
		},
		{
			name: "Should return a valid Version for v2.7.1",
			args: args{
				"v2.7.1",
			},
			want: &Version{
				Major: 2,
				Minor: 7,
				Patch: 1,
			},
			wantErr: false,
		},
		{
			name: "Should fail for an invalid version type for 2.7.1",
			args: args{
				"2.7.1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should fail for an empty string",
			args: args{
				"",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return a valid version type for 2.7.1-beta",
			args: args{
				"v2.7.1-beta",
			},
			want: &Version{
				Major: 2,
				Minor: 7,
				Patch: 1,
				Beta:  true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVRCAutoRejoinTool_GetCurrentVersion(t *testing.T) {

	t.Run("Should return a valid version", func(t *testing.T) {
		v := NewVRCAutoRejoinTool()
		_, err := v.GetCurrentVersion()
		if err != nil {
			t.Errorf("VRCAutoRejoinTool.GetCurrentVersion() error = %v", err)
		}
	})
}
