package stream

import (
	"context"
	"io"
	"testing"

	pb "gRPCServerDemo/gen/stream"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// baseMockStream 提供所有 stream mock 共用的基础方法
type baseMockStream struct{}

func (m *baseMockStream) SetHeader(metadata.MD) error  { return nil }
func (m *baseMockStream) SendHeader(metadata.MD) error  { return nil }
func (m *baseMockStream) SetTrailer(metadata.MD)        {}
func (m *baseMockStream) Context() context.Context       { return context.Background() }
func (m *baseMockStream) SendMsg(interface{}) error      { return nil }
func (m *baseMockStream) RecvMsg(interface{}) error      { return nil }

// --- mock: List ---

type mockListStream struct {
	baseMockStream
	sent []*pb.StreamResponse
}

func (m *mockListStream) Send(resp *pb.StreamResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}

func TestHandler_List(t *testing.T) {
	h := &Handler{}

	t.Run("normal", func(t *testing.T) {
		stream := &mockListStream{}
		req := &pb.StreamRequest{Pt: &pb.StreamPoint{Name: "test", Value: 100}}
		err := h.List(req, stream)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(stream.sent) != 7 {
			t.Fatalf("List() sent %d responses, want 7", len(stream.sent))
		}
		for i, resp := range stream.sent {
			expected := int32(100 + i)
			if resp.Pt.Value != expected {
				t.Errorf("List() response[%d].Value = %d, want %d", i, resp.Pt.Value, expected)
			}
			if resp.Pt.Name != "test" {
				t.Errorf("List() response[%d].Name = %s, want test", i, resp.Pt.Name)
			}
		}
	})

	t.Run("nil pt returns InvalidArgument", func(t *testing.T) {
		stream := &mockListStream{}
		req := &pb.StreamRequest{Pt: nil}
		err := h.List(req, stream)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("List(nil pt) = %v, want InvalidArgument", err)
		}
	})
}

// --- mock: Record ---

type mockRecordStream struct {
	baseMockStream
	received []*pb.StreamRequest
	closed   *pb.StreamResponse
	idx      int
}

func (m *mockRecordStream) SendAndClose(resp *pb.StreamResponse) error {
	m.closed = resp
	return nil
}

func (m *mockRecordStream) Recv() (*pb.StreamRequest, error) {
	if m.idx >= len(m.received) {
		return nil, io.EOF
	}
	req := m.received[m.idx]
	m.idx++
	return req, nil
}

func TestHandler_Record(t *testing.T) {
	h := &Handler{}

	t.Run("normal", func(t *testing.T) {
		stream := &mockRecordStream{
			received: []*pb.StreamRequest{
				{Pt: &pb.StreamPoint{Name: "a", Value: 1}},
				{Pt: &pb.StreamPoint{Name: "b", Value: 2}},
				{Pt: &pb.StreamPoint{Name: "c", Value: 3}},
			},
		}
		err := h.Record(stream)
		if err != nil {
			t.Fatalf("Record() error = %v", err)
		}
		if stream.closed == nil {
			t.Fatal("Record() did not call SendAndClose")
		}
		if stream.closed.Pt.Name != "gRPC Stream Server: Record" {
			t.Errorf("Record() closed.Name = %s, want 'gRPC Stream Server: Record'", stream.closed.Pt.Name)
		}
	})

	t.Run("nil pt returns InvalidArgument", func(t *testing.T) {
		stream := &mockRecordStream{
			received: []*pb.StreamRequest{
				{Pt: nil},
			},
		}
		err := h.Record(stream)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("Record(nil pt) = %v, want InvalidArgument", err)
		}
	})
}

// --- mock: Route ---

type mockRouteStream struct {
	baseMockStream
	sent     []*pb.StreamResponse
	received []*pb.StreamRequest
	idx      int
}

func (m *mockRouteStream) Send(resp *pb.StreamResponse) error {
	m.sent = append(m.sent, resp)
	return nil
}

func (m *mockRouteStream) Recv() (*pb.StreamRequest, error) {
	if m.idx >= len(m.received) {
		return nil, io.EOF
	}
	req := m.received[m.idx]
	m.idx++
	return req, nil
}

func TestHandler_Route(t *testing.T) {
	h := &Handler{}

	t.Run("echo pattern", func(t *testing.T) {
		stream := &mockRouteStream{
			received: []*pb.StreamRequest{
				{Pt: &pb.StreamPoint{Name: "a", Value: 1}},
				{Pt: &pb.StreamPoint{Name: "b", Value: 2}},
				{Pt: &pb.StreamPoint{Name: "c", Value: 3}},
			},
		}
		err := h.Route(stream)
		if err != nil {
			t.Fatalf("Route() error = %v", err)
		}
		if len(stream.sent) != 3 {
			t.Fatalf("Route() sent %d responses, want 3", len(stream.sent))
		}
		for i, resp := range stream.sent {
			expected := int32(i + 1)
			if resp.Pt.Value != expected {
				t.Errorf("Route() response[%d].Value = %d, want %d", i, resp.Pt.Value, expected)
			}
			if resp.Pt.Name != "gRPC Stream Server: Route" {
				t.Errorf("Route() response[%d].Name = %s, want 'gRPC Stream Server: Route'", i, resp.Pt.Name)
			}
		}
	})

	t.Run("nil pt returns InvalidArgument", func(t *testing.T) {
		stream := &mockRouteStream{
			received: []*pb.StreamRequest{
				{Pt: nil},
			},
		}
		err := h.Route(stream)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("Route(nil pt) = %v, want InvalidArgument", err)
		}
	})

	t.Run("empty stream", func(t *testing.T) {
		stream := &mockRouteStream{
			received: []*pb.StreamRequest{},
		}
		err := h.Route(stream)
		if err != nil {
			t.Fatalf("Route(empty) error = %v", err)
		}
		if len(stream.sent) != 0 {
			t.Errorf("Route(empty) sent %d responses, want 0", len(stream.sent))
		}
	})
}
