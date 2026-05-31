package interceptor

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerLoggingInterceptor(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	resp, err := UnaryServerLoggingInterceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got: %v", resp)
	}
}

func TestUnaryServerLoggingInterceptor_Error(t *testing.T) {
	expectedErr := status.Error(codes.Internal, "test error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	_, err := UnaryServerLoggingInterceptor(context.Background(), nil, info, handler)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("Expected Internal code, got: %v", status.Code(err))
	}
}

func TestStreamServerLoggingInterceptor(t *testing.T) {
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	info := &grpc.StreamServerInfo{
		FullMethod: "/test/StreamMethod",
	}

	err := StreamServerLoggingInterceptor(nil, nil, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestStreamServerLoggingInterceptor_Error(t *testing.T) {
	expectedErr := status.Error(codes.Internal, "test error")
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return expectedErr
	}

	info := &grpc.StreamServerInfo{
		FullMethod: "/test/StreamMethod",
	}

	err := StreamServerLoggingInterceptor(nil, nil, info, handler)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("Expected Internal code, got: %v", status.Code(err))
	}
}
