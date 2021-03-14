package main

import (
	"gRPCServerDemo/chat"
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

	s := chat.Server{}
	chat.RegisterChatServiceServer(gRPCServer, &s)
	reflection.Register(gRPCServer)

	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
	}
}
