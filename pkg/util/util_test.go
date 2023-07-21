package util

import "testing"

func TestDecodingBase64(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecodingBase64(tt.args.data); got != tt.want {
				t.Errorf("DecodingBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}
