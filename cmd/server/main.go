package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	genadmin "gRPCServerDemo/gen/admin"
	genchat "gRPCServerDemo/gen/chat"
	genstream "gRPCServerDemo/gen/stream"
	"gRPCServerDemo/internal/admin"
	"gRPCServerDemo/internal/chat"
	"gRPCServerDemo/internal/interceptor"
	"gRPCServerDemo/internal/stream"
)

func main() {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}

	gRPCServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryServerRecoveryInterceptor,
			interceptor.UnaryServerLoggingInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptor.StreamServerRecoveryInterceptor,
			interceptor.StreamServerLoggingInterceptor,
		),
	)

	// 注册业务服务
	genchat.RegisterChatServiceServer(gRPCServer, &chat.Handler{})
	genstream.RegisterStreamServiceServer(gRPCServer, &stream.Handler{})

	// 注册AdminService（在其他服务之后注册）
	genadmin.RegisterAdminServiceServer(gRPCServer, admin.NewHandler(gRPCServer))

	// 注册健康检查服务（全局 + 各服务）
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(gRPCServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("proto.ChatService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("proto.StreamService", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("proto.AdminService", healthpb.HealthCheckResponse_SERVING)

	// 注册反射服务（方便 grpcurl 调试）
	reflection.Register(gRPCServer)

	// 优雅关闭：监听系统信号
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("Received signal: %v, shutting down gracefully...", sig)

		// 标记所有服务为未就绪
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("proto.ChatService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("proto.StreamService", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("proto.AdminService", healthpb.HealthCheckResponse_NOT_SERVING)

		// 带超时的优雅关闭，超时后强制关闭
		done := make(chan struct{})
		go func() {
			gRPCServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Server stopped gracefully")
		case <-time.After(5 * time.Second):
			log.Println("Graceful stop timed out, forcing stop")
			gRPCServer.Stop()
		}
	}()

	log.Println("gRPC server is running on :9000")
	if err := gRPCServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}