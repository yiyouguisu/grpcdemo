package interceptor

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"gRPCServerDemo/internal/auditlog"
)

// timeNow returns today's date formatted as "2006-01-02", matching the
// log file naming convention used by FileManager.
func timeNow() string {
	return time.Now().Format("2006-01-02")
}

// unserializable is a type that cannot be serialized to JSON,
// used to test serialization failure handling.
type unserializable func()

// mockServerStreamForAudit is a minimal grpc.ServerStream mock for testing
// the audit stream interceptor.
type mockServerStreamForAudit struct {
	grpc.ServerStream
	ctx      context.Context
	recvFunc func(m any) error
}

func (m *mockServerStreamForAudit) Context() context.Context {
	return m.ctx
}

func (m *mockServerStreamForAudit) RecvMsg(msg any) error {
	if m.recvFunc != nil {
		return m.recvFunc(msg)
	}
	return io.EOF
}

func (m *mockServerStreamForAudit) SendMsg(msg any) error {
	return nil
}

func (m *mockServerStreamForAudit) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStreamForAudit) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockServerStreamForAudit) SetTrailer(metadata.MD) {
}

func Test_extractServiceName(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
		want       string
	}{
		{"standard", "/ChatService/SendMessage", "ChatService"},
		{"admin", "/AdminService/ListServices", "AdminService"},
		{"stream", "/StreamService/List", "StreamService"},
		{"empty", "", ""},
		{"no service", "/Method", "Method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractServiceName(tt.fullMethod)
			if got != tt.want {
				t.Errorf("extractServiceName(%q) = %q, want %q", tt.fullMethod, got, tt.want)
			}
		})
	}
}

func TestNewAuditUnaryInterceptor_Success(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	interceptor := NewAuditUnaryInterceptor(manager)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/ChatService/SayHello"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got: %v", resp)
	}

	// Verify log file was created and contains expected content.
	logFile := filepath.Join(logDir, "ChatService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "/ChatService/SayHello") {
		t.Errorf("Log should contain method name, got: %s", content)
	}
	if !strings.Contains(content, "request") {
		t.Errorf("Log should contain request data, got: %s", content)
	}
	if !strings.Contains(content, "response") {
		t.Errorf("Log should contain response data, got: %s", content)
	}
}

func TestNewAuditUnaryInterceptor_Error(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	interceptor := NewAuditUnaryInterceptor(manager)

	expectedErr := status.Error(codes.NotFound, "not found")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/AdminService/ListServices"}

	_, err := interceptor(context.Background(), "request", info, handler)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("Expected NotFound code, got: %v", status.Code(err))
	}

	// Verify log file contains error information.
	logFile := filepath.Join(logDir, "AdminService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "/AdminService/ListServices") {
		t.Errorf("Log should contain method name, got: %s", content)
	}
	if !strings.Contains(content, "request") {
		t.Errorf("Log should contain request data, got: %s", content)
	}
	if !strings.Contains(content, "not found") {
		t.Errorf("Log should contain error message, got: %s", content)
	}
}

func TestNewAuditUnaryInterceptor_SerializationError(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	interceptor := NewAuditUnaryInterceptor(manager)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/ChatService/SayHello"}

	// Pass an unserializable request (func type); interceptor should still succeed.
	var unserializableReq unserializable = func() {}
	resp, err := interceptor(context.Background(), &unserializableReq, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp != "ok" {
		t.Errorf("Expected 'ok', got: %v", resp)
	}

	// Verify log file was created and contains an error marker for the request.
	logFile := filepath.Join(logDir, "ChatService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "/ChatService/SayHello") {
		t.Errorf("Log should contain method name, got: %s", content)
	}
	if !strings.Contains(content, "error") {
		t.Errorf("Log should contain serialization error, got: %s", content)
	}
}

func TestNewAuditStreamInterceptor_RecordsStartAndEnd(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	streamInterceptor := NewAuditStreamInterceptor(manager)

	mockSS := &mockServerStreamForAudit{
		ctx: context.Background(),
	}

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}
	info := &grpc.StreamServerInfo{FullMethod: "/StreamService/List"}

	err := streamInterceptor(nil, mockSS, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify log file was created and contains stream_ended.
	logFile := filepath.Join(logDir, "StreamService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "/StreamService/List") {
		t.Errorf("Log should contain method name, got: %s", content)
	}
	if !strings.Contains(content, "stream_ended") {
		t.Errorf("Log should contain stream_ended event, got: %s", content)
	}
}

func TestNewAuditStreamInterceptor_Error(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	streamInterceptor := NewAuditStreamInterceptor(manager)

	mockSS := &mockServerStreamForAudit{
		ctx: context.Background(),
	}

	expectedErr := status.Error(codes.Internal, "stream error")
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return expectedErr
	}
	info := &grpc.StreamServerInfo{FullMethod: "/StreamService/Route"}

	err := streamInterceptor(nil, mockSS, info, handler)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("Expected Internal code, got: %v", status.Code(err))
	}

	// Verify log file contains error information.
	logFile := filepath.Join(logDir, "StreamService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "stream error") {
		t.Errorf("Log should contain error message, got: %s", content)
	}
	if !strings.Contains(content, "Internal") {
		t.Errorf("Log should contain error code, got: %s", content)
	}
}

func TestNewAuditStreamInterceptor_WithRecvMsg(t *testing.T) {
	logDir := t.TempDir()
	config := auditlog.AuditLogConfig{LogDir: logDir, RetainDays: 7}
	manager := auditlog.NewFileManager(config)
	manager.Start()
	defer manager.Close()

	streamInterceptor := NewAuditStreamInterceptor(manager)

	// Create a mock stream where RecvMsg succeeds once then returns io.EOF.
	recvCount := 0
	mockSS := &mockServerStreamForAudit{
		ctx: context.Background(),
		recvFunc: func(m any) error {
			recvCount++
			if recvCount == 1 {
				return nil
			}
			return io.EOF
		},
	}

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		// Simulate receiving one message then ending.
		var msg interface{}
		_ = stream.RecvMsg(msg)
		_ = stream.RecvMsg(msg) // returns EOF
		return nil
	}
	info := &grpc.StreamServerInfo{FullMethod: "/StreamService/Route"}

	err := streamInterceptor(nil, mockSS, info, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	logFile := filepath.Join(logDir, "StreamService", timeNow()+".log")
	data, readErr := os.ReadFile(logFile)
	if readErr != nil {
		t.Fatalf("Failed to read log file: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "stream_started") {
		t.Errorf("Log should contain stream_started event, got: %s", content)
	}
	if !strings.Contains(content, "stream_ended") {
		t.Errorf("Log should contain stream_ended event, got: %s", content)
	}
}
