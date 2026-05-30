package interceptor

import (
	"context"
	"log"
	"runtime/debug"
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

// UnaryServerRecoveryInterceptor 捕获一元 RPC 中的 panic，返回 Internal 错误
func UnaryServerRecoveryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Recovery] panic recovered in %s: %v\n%s", info.FullMethod, r, debug.Stack())
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}

// StreamServerRecoveryInterceptor 捕获流式 RPC 中的 panic，返回 Internal 错误
func StreamServerRecoveryInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Recovery] panic recovered in %s: %v\n%s", info.FullMethod, r, debug.Stack())
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(srv, ss)
}
