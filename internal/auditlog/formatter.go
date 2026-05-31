package auditlog

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	// timestampFormat uses RFC3339 with millisecond precision.
	timestampFormat = "2006-01-02T15:04:05.000Z07:00"
)

// FormatLogEntry formats a single gRPC call audit log entry as a human-readable
// single-line text record. The output format is:
//
//	[timestamp] methodName requestJSON responseJSON
//
// If err is non-nil, the response section contains the error message instead
// of the response body. If req or resp is nil, it is serialized as "null".
func FormatLogEntry(methodName string, req, resp interface{}, err error) string {
	timestamp := time.Now().Format(timestampFormat)

	reqJSON := "null"
	if req != nil {
		if s, marshalErr := protoToJSON(req); marshalErr == nil {
			reqJSON = s
		} else {
			reqJSON = fmt.Sprintf(`{"error":"%s"}`, marshalErr.Error())
		}
	}

	var respJSON string
	if err != nil {
		respJSON = fmt.Sprintf(`{"error":"%s"}`, err.Error())
	} else if resp != nil {
		if s, marshalErr := protoToJSON(resp); marshalErr == nil {
			respJSON = s
		} else {
			respJSON = fmt.Sprintf(`{"error":"%s"}`, marshalErr.Error())
		}
	} else {
		respJSON = "null"
	}

	return fmt.Sprintf("[%s] %s %s %s", timestamp, methodName, reqJSON, respJSON)
}

// protoToJSON converts a message to a JSON string. If the message implements
// proto.Message, it uses protojson for canonical proto-aware serialization.
// For other types, it falls back to encoding/json. Returns an error string
// wrapped in a JSON object if serialization fails.
func protoToJSON(msg interface{}) (string, error) {
	if msg == nil {
		return "null", nil
	}

	if pm, ok := msg.(proto.Message); ok {
		data, err := protojson.Marshal(pm)
		if err != nil {
			return "", fmt.Errorf("proto marshal: %w", err)
		}
		return string(data), nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(data), nil
}
