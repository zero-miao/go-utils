package grpc_error

import "testing"

func TestRegisterError(t *testing.T) {
	type args struct {
		item ErrorType
	}
	tests := []struct {
		name string
		args args
	}{
		{"", args{item: ErrorType{Code: ENetwork}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterError(tt.args.item)
		})
	}
}
