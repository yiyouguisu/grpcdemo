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

- `proto/` — protobuf 定义：chat.proto (一元)、stream.proto (三种流式)、admin.proto (管理)
- `gen/` — protoc 生成的代码，不要手动修改
- `internal/chat/` — ChatService 实现 (SayHello 一元 RPC)
- `internal/stream/` — StreamService 实现 (List/Record/Route 流式 RPC)
- `internal/admin/` — AdminService 实现 (ListServices 查询服务列表)
- `internal/interceptor/` — 日志和 panic 恢复拦截器
- `cmd/server/` — 服务端入口，注册服务、健康检查、反射、优雅关闭
- `cmd/client/` — 客户端入口，演示所有 RPC 调用方式

## 编码规范

详见 [CODING_STYLE.md](CODING_STYLE.md)，关键要点：

- import 分组顺序：标准库、第三方库、内部包
- 服务实现统一使用 `handler.go` + `Handler` 结构体
- proto 文件统一使用 4 空格缩进
- 测试使用 baseMockStream 嵌入结构体减少重复代码
- 魔法数字提取为常量 (如 streamCount)
