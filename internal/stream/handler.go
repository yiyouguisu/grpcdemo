package stream

import (
	"io"
	"log"

	pb "gRPCServerDemo/gen/stream"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const streamCount = 6

type Handler struct {
	pb.UnimplementedStreamServiceServer
}

func (h *Handler) List(r *pb.StreamRequest, stream pb.StreamService_ListServer) error {
	if r.Pt == nil {
		return status.Error(codes.InvalidArgument, "pt is required")
	}
	for n := 0; n <= streamCount; n++ {
		err := stream.Send(&pb.StreamResponse{
			Pt: &pb.StreamPoint{
				Name:  r.Pt.Name,
				Value: r.Pt.Value + int32(n),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) Record(stream pb.StreamService_RecordServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.StreamResponse{Pt: &pb.StreamPoint{Name: "gRPC Stream Server: Record", Value: 1}})
		}
		if err != nil {
			return err
		}
		if r.Pt == nil {
			return status.Error(codes.InvalidArgument, "pt is required")
		}
		log.Printf("stream.Recv pt.name: %s, pt.value: %d", r.Pt.Name, r.Pt.Value)
	}
}

func (h *Handler) Route(stream pb.StreamService_RouteServer) error {
	for {
		// 先接收客户端消息
		r, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if r.Pt == nil {
			return status.Error(codes.InvalidArgument, "pt is required")
		}

		log.Printf("stream.Recv pt.name: %s, pt.value: %d", r.Pt.Name, r.Pt.Value)

		// 再发送响应
		err = stream.Send(&pb.StreamResponse{
			Pt: &pb.StreamPoint{
				Name:  "gRPC Stream Server: Route",
				Value: r.Pt.Value,
			},
		})
		if err != nil {
			return err
		}
	}
}
