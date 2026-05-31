package interceptor

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerRecoveryInterceptor(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	resp, err := UnaryServerRecoveryInterceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got: %v", resp)
	}
}

func TestUnaryServerRecoveryInterceptor_Panic(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("test panic")
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	_, err := UnaryServerRecoveryInterceptor(context.Background(), nil, info, handler)
	if err == nil {
		t.Fatal("Expected error from panic recovery, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("Expected Internal code, got: %v", status.Code(err))
	}
}

func TestStreamServerRecoveryInterceptor(t *testing.T) {
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	info := &grpc.StreamServerInfo{
		FullMethod: "/test/StreamMethod",
	}

	err := StreamServerRecoveryInterceptor(nil, nil, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestStreamServerRecoveryInterceptor_Panic(t *testing.T) {
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		panic("test panic")
	}

	info := &grpc.StreamServerInfo{
		FullMethod: "/test/StreamMethod",
	}

	err := StreamServerRecoveryInterceptor(nil, nil, info, handler)
	if err == nil {
		t.Fatal("Expected error from panic recovery, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("Expected Internal code, got: %v", status.Code(err))
	}
}
