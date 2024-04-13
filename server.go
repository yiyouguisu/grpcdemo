package main

import (
	"gRPCServerDemo/chat"
	"gRPCServerDemo/stream"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}

	gRPCServer := grpc.NewServer()
	chat.RegisterChatServiceServer(gRPCServer, &chat.Server{})
	stream.RegisterStreamServiceServer(gRPCServer, &stream.StreamService{})
	reflection.Register(gRPCServer)

	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
	}
}
