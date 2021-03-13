package chat

import (
	"context"
	"reflect"
	"testing"
)

func TestServer_SayHello(t *testing.T) {
	type args struct {
		ctx     context.Context
		message *Message
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *Message
		wantErr bool
	}{
		{
			name: "test one",
			s:    &Server{},
			args: args{
				ctx:     context.Background(),
				message: &Message{Body: "hello"},
			},
			want:    &Message{Body: "Hello from the server!"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SayHello(tt.args.ctx, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.SayHello() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.SayHello() = %v, want %v", got, tt.want)
			}
		})
	}
}
