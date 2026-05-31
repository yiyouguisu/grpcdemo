package admin

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestHandler_ListServices(t *testing.T) {
	server := grpc.NewServer()
	handler := NewHandler(server)

	resp, err := handler.ListServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	if resp.ServiceNames == nil {
		t.Error("Expected service names, got nil")
	}

	t.Logf("Services: %v", resp.ServiceNames)
}

func TestNewHandler(t *testing.T) {
	server := grpc.NewServer()
	handler := NewHandler(server)

	if handler == nil {
		t.Fatal("Expected non-nil Handler")
	}

	if handler.server != server {
		t.Error("Expected server to be set correctly")
	}
}
