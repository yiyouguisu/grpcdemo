package main

import (
	"context"
	"gRPCServerDemo/chat"
	"log"

	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connet, %s", err)
	}
	defer conn.Close()

	c := chat.NewChatServiceClient(conn)
	message := &chat.RequstMessage{
		Body:   "Hello from the client!",
	}
	response, err := c.SayHello(context.Background(), message)
	if err != nil {
		log.Fatalf("Error when calling SayHello, %s", err)
	}
	log.Printf("Response from Server: %v", response.Body)
}
