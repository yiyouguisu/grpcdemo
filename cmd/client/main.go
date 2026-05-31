package main

import (
	"context"
	"fmt"
	"io"
	"log"

	adminpb "gRPCServerDemo/gen/admin"
	chatpb "gRPCServerDemo/gen/chat"
	streampb "gRPCServerDemo/gen/stream"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const streamCount = 6

func main() {
	conn, err := grpc.NewClient("passthrough:///localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("=== Unary RPC: SayHello ===")
	testSayHello(conn)

	fmt.Println("\n=== Server Streaming: List ===")
	testList(conn)

	fmt.Println("\n=== Client Streaming: Record ===")
	testRecord(conn)

	fmt.Println("\n=== Bidirectional Streaming: Route ===")
	testRoute(conn)

	fmt.Println("\n=== Unary RPC: ListServices ===")
	testListServices(conn)
}

// 一元 RPC
func testSayHello(conn *grpc.ClientConn) {
	client := chatpb.NewChatServiceClient(conn)
	resp, err := client.SayHello(context.Background(), &chatpb.RequestMessage{Body: "Hello from the client!"})
	if err != nil {
		log.Fatalf("SayHello error: %v", err)
	}
	log.Printf("Response: %s", resp.Body)
}

// 服务端流式 RPC
func testList(conn *grpc.ClientConn) {
	client := streampb.NewStreamServiceClient(conn)
	stream, err := client.List(context.Background(), &streampb.StreamRequest{
		Pt: &streampb.StreamPoint{Name: "gRPC Stream Client: List", Value: 2018},
	})
	if err != nil {
		log.Fatalf("List error: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("List recv error: %v", err)
		}
		log.Printf("resp: pt.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)
	}
}

// 客户端流式 RPC
func testRecord(conn *grpc.ClientConn) {
	client := streampb.NewStreamServiceClient(conn)
	stream, err := client.Record(context.Background())
	if err != nil {
		log.Fatalf("Record error: %v", err)
	}

	for n := 0; n < streamCount; n++ {
		err := stream.Send(&streampb.StreamRequest{
			Pt: &streampb.StreamPoint{Name: "gRPC Stream Client: Record", Value: int32(n)},
		})
		if err != nil {
			log.Fatalf("Record send error: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Record CloseAndRecv error: %v", err)
	}
	log.Printf("resp: pt.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)
}

// 双向流式 RPC
func testRoute(conn *grpc.ClientConn) {
	client := streampb.NewStreamServiceClient(conn)
	stream, err := client.Route(context.Background())
	if err != nil {
		log.Fatalf("Route error: %v", err)
	}

	for n := 0; n <= streamCount; n++ {
		err := stream.Send(&streampb.StreamRequest{
			Pt: &streampb.StreamPoint{Name: "gRPC Stream Client: Route", Value: int32(n)},
		})
		if err != nil {
			log.Fatalf("Route send error: %v", err)
		}

		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Route recv error: %v", err)
		}
		log.Printf("resp: pt.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)
	}

	// 关闭发送，排空服务端剩余响应
	stream.CloseSend()
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Route drain error: %v", err)
		}
		log.Printf("resp: pt.name: %s, pt.value: %d", resp.Pt.Name, resp.Pt.Value)
	}
}

// Admin 服务：列出所有注册的服务
func testListServices(conn *grpc.ClientConn) {
	client := adminpb.NewAdminServiceClient(conn)
	resp, err := client.ListServices(context.Background(), &adminpb.ListServicesRequest{})
	if err != nil {
		log.Fatalf("ListServices error: %v", err)
	}
	log.Printf("Registered services: %v", resp.ServiceNames)
}
