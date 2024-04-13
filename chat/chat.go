package chat

import (
	"context"
	"log"
)

type Server struct {
	UnimplementedChatServiceServer
}

func (s *Server) SayHello(ctx context.Context, message *RequstMessage) (*ResponseMessage, error) {
	log.Printf("Received message body from client, %s", message.Body)
	return &ResponseMessage{
		Body: "Hello from the server!",
	}, nil
}
