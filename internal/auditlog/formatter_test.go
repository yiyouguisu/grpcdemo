package auditlog

import (
	"errors"
	"strings"
	"testing"
	"time"

	genchat "gRPCServerDemo/gen/chat"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFormatLogEntry_StandardFormat(t *testing.T) {
	req := &genchat.RequestMessage{Body: "hello"}
	resp := &genchat.ResponseMessage{Body: "world"}

	entry := FormatLogEntry("/ChatService/SayHello", req, resp, nil)

	// Should start with a timestamp in brackets.
	if !strings.HasPrefix(entry, "[") {
		t.Errorf("expected entry to start with '[', got %q", entry)
	}

	// Should contain the method name.
	if !strings.Contains(entry, "/ChatService/SayHello") {
		t.Errorf("expected entry to contain method name, got %q", entry)
	}

	// Should contain request JSON body.
	if !strings.Contains(entry, `"body":"hello"`) {
		t.Errorf("expected entry to contain request body, got %q", entry)
	}

	// Should contain response JSON body.
	if !strings.Contains(entry, `"body":"world"`) {
		t.Errorf("expected entry to contain response body, got %q", entry)
	}

	// Should be a single line.
	if strings.Contains(entry, "\n") {
		t.Errorf("expected single-line entry, got %q", entry)
	}
}

func TestFormatLogEntry_NilRequest(t *testing.T) {
	resp := &genchat.ResponseMessage{Body: "world"}

	entry := FormatLogEntry("/ChatService/SayHello", nil, resp, nil)

	if !strings.Contains(entry, "null") {
		t.Errorf("expected 'null' for nil request, got %q", entry)
	}
}

func TestFormatLogEntry_NilResponse(t *testing.T) {
	req := &genchat.RequestMessage{Body: "hello"}

	entry := FormatLogEntry("/ChatService/SayHello", req, nil, nil)

	// After the method name and request, response section should be "null".
	parts := strings.SplitN(entry, " ", 4)
	if len(parts) < 4 {
		t.Fatalf("expected at least 4 parts, got %d: %q", len(parts), entry)
	}
	if parts[3] != "null" {
		t.Errorf("expected response section 'null', got %q", parts[3])
	}
}

func TestFormatLogEntry_WithError(t *testing.T) {
	req := &genchat.RequestMessage{Body: "hello"}
	rpcErr := status.Errorf(codes.Internal, "something went wrong")

	entry := FormatLogEntry("/AdminService/ListServices", req, nil, rpcErr)

	// Should contain the error message.
	if !strings.Contains(entry, "something went wrong") {
		t.Errorf("expected error message in entry, got %q", entry)
	}

	// Should contain error JSON structure.
	if !strings.Contains(entry, `"error":`) {
		t.Errorf("expected error JSON in entry, got %q", entry)
	}

	// Should NOT contain "null" for the response (error replaces it).
	parts := strings.SplitN(entry, " ", 4)
	if len(parts) < 4 {
		t.Fatalf("expected at least 4 parts, got %d: %q", len(parts), entry)
	}
	if parts[3] == "null" {
		t.Error("expected error in response section, got 'null'")
	}
}

func TestFormatLogEntry_RequestSerializationFailure(t *testing.T) {
	// Use an unserializable type (a channel) to trigger serialization error.
	req := make(chan int)
	resp := &genchat.ResponseMessage{Body: "world"}

	entry := FormatLogEntry("/ChatService/SayHello", req, resp, nil)

	// Should contain serialization error info in request section.
	if !strings.Contains(entry, "error") {
		t.Errorf("expected serialization error in entry, got %q", entry)
	}

	// The entry should still be a single line and not panic.
	if strings.Contains(entry, "\n") {
		t.Errorf("expected single-line entry, got %q", entry)
	}
}

func TestFormatLogEntry_ResponseSerializationFailure(t *testing.T) {
	req := &genchat.RequestMessage{Body: "hello"}
	resp := make(chan int) // unserializable

	entry := FormatLogEntry("/ChatService/SayHello", req, resp, nil)

	// Should contain serialization error info in response section.
	if !strings.Contains(entry, "error") {
		t.Errorf("expected serialization error in entry, got %q", entry)
	}
}

func TestFormatLogEntry_BothNilNoError(t *testing.T) {
	entry := FormatLogEntry("/ChatService/SayHello", nil, nil, nil)

	if !strings.Contains(entry, "/ChatService/SayHello") {
		t.Errorf("expected method name, got %q", entry)
	}

	// Both request and response should be "null".
	parts := strings.SplitN(entry, " ", 4)
	if len(parts) < 4 {
		t.Fatalf("expected at least 4 parts, got %d: %q", len(parts), entry)
	}
	if parts[2] != "null" || parts[3] != "null" {
		t.Errorf("expected both req and resp as 'null', got req=%q resp=%q", parts[2], parts[3])
	}
}

func TestFormatLogEntry_TimestampFormat(t *testing.T) {
	entry := FormatLogEntry("/Test/Method", nil, nil, nil)

	// Extract timestamp between first [ and ]
	start := strings.Index(entry, "[")
	end := strings.Index(entry, "]")
	if start < 0 || end < 0 || end <= start {
		t.Fatalf("expected [timestamp] format, got %q", entry)
	}

	ts := entry[start+1 : end]
	_, err := time.Parse(timestampFormat, ts)
	if err != nil {
		t.Errorf("timestamp %q does not match format %s: %v", ts, timestampFormat, err)
	}
}

func TestFormatLogEntry_NonProtoMessage(t *testing.T) {
	req := map[string]string{"key": "value"}
	resp := map[string]int{"count": 42}

	entry := FormatLogEntry("/Test/Method", req, resp, nil)

	if !strings.Contains(entry, `"key":"value"`) {
		t.Errorf("expected map request in entry, got %q", entry)
	}
	if !strings.Contains(entry, `"count":42`) {
		t.Errorf("expected map response in entry, got %q", entry)
	}
}

func TestFormatLogEntry_PlainError(t *testing.T) {
	req := &genchat.RequestMessage{Body: "hello"}
	plainErr := errors.New("connection refused")

	entry := FormatLogEntry("/ChatService/SayHello", req, nil, plainErr)

	if !strings.Contains(entry, "connection refused") {
		t.Errorf("expected plain error message, got %q", entry)
	}
}

func TestFormatLogEntry_ErrorWithResponse(t *testing.T) {
	// When err is non-nil, resp should be ignored.
	req := &genchat.RequestMessage{Body: "hello"}
	resp := &genchat.ResponseMessage{Body: "world"}
	rpcErr := status.Errorf(codes.NotFound, "not found")

	entry := FormatLogEntry("/Test/Method", req, resp, rpcErr)

	// Response section should contain error, not the response body.
	parts := strings.SplitN(entry, " ", 4)
	if len(parts) < 4 {
		t.Fatalf("expected at least 4 parts, got %d: %q", len(parts), entry)
	}
	if !strings.Contains(parts[3], "not found") {
		t.Errorf("expected error in response section, got %q", parts[3])
	}
	if strings.Contains(parts[3], "world") {
		t.Errorf("response body should not appear when error is present, got %q", parts[3])
	}
}

func TestProtoToJSON_ProtoMessage(t *testing.T) {
	msg := &genchat.RequestMessage{Body: "test"}
	s, err := protoToJSON(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s, `"test"`) {
		t.Errorf("expected JSON to contain 'test', got %q", s)
	}
}

func TestProtoToJSON_NilMessage(t *testing.T) {
	s, err := protoToJSON(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "null" {
		t.Errorf("expected 'null', got %q", s)
	}
}

func TestProtoToJSON_NonProtoMessage(t *testing.T) {
	msg := map[string]string{"a": "b"}
	s, err := protoToJSON(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s, `"a":"b"`) {
		t.Errorf("expected JSON to contain key-value, got %q", s)
	}
}

func TestProtoToJSON_UnserializableType(t *testing.T) {
	msg := make(chan int)
	_, err := protoToJSON(msg)
	if err == nil {
		t.Error("expected error for unserializable type")
	}
}
