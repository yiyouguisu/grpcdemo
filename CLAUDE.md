# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

gRPC 演示项目，展示四种通信模式：一元 RPC、服务端流、客户端流、双向流。

## 常用命令

```bash
make run        # 启动 gRPC 服务端 (:9000)
make client     # 运行客户端测试所有 RPC 模式
make build      # 编译到 bin/
make test       # 运行测试 (含 race 检测)
make proto      # 重新生成 protobuf 代码
make lint       # golangci-lint 检查
```

## 架构

- `proto/` — protobuf 定义，`chat.proto` (一元) 和 `stream.proto` (三种流式)
- `gen/` — protoc 生成的代码，不要手动修改
- `internal/chat/` — ChatService 实现 (SayHello 一元 RPC)
- `internal/stream/` — StreamService 实现 (List/Record/Route 流式 RPC)
- `internal/interceptor/` — 日志和 panic 恢复拦截器
- `cmd/server/` — 服务端入口，注册服务、健康检查、反射、优雅关闭
- `cmd/client/` — 客户端入口，演示四种 RPC 调用方式

## 注意事项

- Go module 名为 `gRPCServerDemo`，不是仓库名
- `RequstMessage` 是已知的拼写错误 (proto 第 8 行注释说明)
- 生成代码在 `gen/` 而非传统的 `pb/` 目录
- 服务默认监听 `:9000`
