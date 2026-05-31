package chat

import (
	"context"
	"log"

	pb "gRPCServerDemo/gen/chat"
)

type Handler struct {
	pb.UnimplementedChatServiceServer
}

func (h *Handler) SayHello(ctx context.Context, message *pb.RequestMessage) (*pb.ResponseMessage, error) {
	log.Printf("Received message body from client: %s", message.Body)
	return &pb.ResponseMessage{
		Body: "Hello from the server!",
	}, nil
}
