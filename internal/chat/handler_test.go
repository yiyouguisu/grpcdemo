package chat

import (
	"context"
	"testing"

	pb "gRPCServerDemo/gen/chat"
)

func TestHandler_SayHello(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{
			name: "normal request",
			body: "hello",
			want: "Hello from the server!",
		},
		{
			name: "empty body",
			body: "",
			want: "Hello from the server!",
		},
	}

	h := &Handler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.SayHello(context.Background(), &pb.RequestMessage{Body: tt.body})
			if (err != nil) != tt.wantErr {
				t.Errorf("SayHello() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Body != tt.want {
				t.Errorf("SayHello() = %v, want %v", got.Body, tt.want)
			}
		})
	}
}
