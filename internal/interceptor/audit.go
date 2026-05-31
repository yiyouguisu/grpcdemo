package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"gRPCServerDemo/internal/auditlog"
)

// auditTimestampFormat uses RFC3339 with millisecond precision, matching
// the format used by auditlog.FormatLogEntry.
const auditTimestampFormat = "2006-01-02T15:04:05.000Z07:00"

// extractServiceName parses the service name from a gRPC full method path.
// For example, "/ChatService/SendMessage" returns "ChatService".
func extractServiceName(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// NewAuditUnaryInterceptor returns a grpc.UnaryServerInterceptor that records
// the complete request and response for every unary RPC call. Log entries are
// written to per-service, date-stamped files via the provided FileManager.
//
// Log write failures are logged to stderr but never affect the original request.
func NewAuditUnaryInterceptor(manager *auditlog.FileManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		serviceName := extractServiceName(info.FullMethod)
		entry := auditlog.FormatLogEntry(info.FullMethod, req, resp, err)

		writer, writerErr := manager.GetLogWriter(serviceName)
		if writerErr != nil {
			log.Printf("[Audit] failed to get log writer for %s: %v", serviceName, writerErr)
			return resp, err
		}

		if _, writeErr := io.WriteString(writer, entry+"\n"); writeErr != nil {
			log.Printf("[Audit] failed to write log for %s: %v", info.FullMethod, writeErr)
		}

		return resp, err
	}
}

// auditServerStream wraps a grpc.ServerStream to capture the initial request
// message on the first successful RecvMsg call. The captured message is
// serialized to JSON and passed to the onFirstRecv callback. Subsequent
// RecvMsg calls are passed through without interception.
type auditServerStream struct {
	grpc.ServerStream
	onFirstRecv func(string)
	recvOnce    sync.Once
}

// RecvMsg intercepts the first successful receive to capture and serialize
// the initial request message. The callback fires only once.
func (s *auditServerStream) RecvMsg(m any) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.recvOnce.Do(func() {
			reqJSON := serializeMessage(m)
			if s.onFirstRecv != nil {
				s.onFirstRecv(reqJSON)
			}
		})
	}
	return err
}

// NewAuditStreamInterceptor returns a grpc.StreamServerInterceptor that records
// stream establishment and closure for every streaming RPC call. The initial
// request message is captured via a wrapped stream and logged when first received.
// Intermediate messages are not recorded.
//
// Log write failures are logged to stderr but never affect the original stream handling.
func NewAuditStreamInterceptor(manager *auditlog.FileManager) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		serviceName := extractServiceName(info.FullMethod)
		startTime := time.Now()

		// Wrap the stream to capture the initial request on first RecvMsg.
		wrapped := &auditServerStream{
			ServerStream: ss,
			onFirstRecv: func(reqJSON string) {
				entry := fmt.Sprintf("[%s] %s %s %s",
					startTime.Format(auditTimestampFormat),
					info.FullMethod,
					reqJSON,
					`{"event":"stream_started"}`)
				writeLogEntry(manager, serviceName, entry)
			},
		}

		// Execute the actual stream handler.
		err := handler(srv, wrapped)

		// Log stream closure.
		endTime := time.Now()
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}
		durationMs := endTime.Sub(startTime).Milliseconds()

		var endEntry string
		if err != nil {
			endEntry = fmt.Sprintf("[%s] %s null {\"event\":\"stream_ended\",\"code\":\"%s\",\"error\":\"%s\",\"duration_ms\":%d}",
				endTime.Format(auditTimestampFormat), info.FullMethod, code, err.Error(), durationMs)
		} else {
			endEntry = fmt.Sprintf("[%s] %s null {\"event\":\"stream_ended\",\"code\":\"%s\",\"duration_ms\":%d}",
				endTime.Format(auditTimestampFormat), info.FullMethod, code, durationMs)
		}
		writeLogEntry(manager, serviceName, endEntry)

		return err
	}
}

// serializeMessage converts a message to a JSON string. If the message
// implements proto.Message, it uses protojson for canonical proto-aware
// serialization. For other types, it falls back to encoding/json.
// Serialization errors are wrapped in a JSON error object.
func serializeMessage(msg interface{}) string {
	if msg == nil {
		return "null"
	}
	if pm, ok := msg.(proto.Message); ok {
		data, err := protojson.Marshal(pm)
		if err != nil {
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}
		return string(data)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(data)
}

// writeLogEntry is a helper that writes a pre-formatted log entry to the
// appropriate service log file via the FileManager. Write failures are logged
// to stderr but never propagated to the caller.
func writeLogEntry(manager *auditlog.FileManager, serviceName, entry string) {
	writer, err := manager.GetLogWriter(serviceName)
	if err != nil {
		log.Printf("[Audit] failed to get log writer for %s: %v", serviceName, err)
		return
	}
	if _, err := io.WriteString(writer, entry+"\n"); err != nil {
		log.Printf("[Audit] failed to write log entry for %s: %v", serviceName, err)
	}
}