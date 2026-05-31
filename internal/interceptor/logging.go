package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerLoggingInterceptor 记录一元 RPC 的请求方法、耗时和状态码
func UnaryServerLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	code := codes.OK
	if err != nil {
		code = status.Code(err)
	}

	log.Printf("[Unary] method=%s duration=%s code=%s", info.FullMethod, duration, code)
	return resp, err
}

// StreamServerLoggingInterceptor 记录流式 RPC 的请求方法、耗时和状态码
func StreamServerLoggingInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)
	duration := time.Since(start)

	code := codes.OK
	if err != nil {
		code = status.Code(err)
	}

	log.Printf("[Stream] method=%s duration=%s code=%s", info.FullMethod, duration, code)
	return err
}
