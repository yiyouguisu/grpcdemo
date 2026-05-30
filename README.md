# grpcdemo

一个 Go gRPC 学习/演示项目，展示了 gRPC 的全部四种通信模式。

## 项目结构

```
grpcdemo/
├── cmd/
│   ├── server/          # 服务端入口
│   │   └── main.go
│   └── client/          # 客户端入口
│       └── main.go
├── internal/
│   ├── chat/            # ChatService 实现（一元 RPC）
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── stream/          # StreamService 实现（流式 RPC）
│   │   ├── handler.go
│   │   └── handler_test.go
│   └── interceptor/     # gRPC 拦截器
│       └── interceptor.go
├── gen/                 # protobuf 生成的代码
│   ├── chat/
│   └── stream/
├── proto/               # protobuf 定义文件
│   ├── chat.proto
│   └── stream.proto
├── Makefile
├── Dockerfile
└── go.mod
```

## 四种 gRPC 通信模式

| 模式 | RPC 方法 | 说明 |
|------|---------|------|
| **一元 RPC** | `ChatService.SayHello` | 客户端发一条消息，服务端返回一条 |
| **服务端流** | `StreamService.List` | 客户端发一条，服务端返回多条 |
| **客户端流** | `StreamService.Record` | 客户端发多条，服务端返回一条 |
| **双向流** | `StreamService.Route` | 双方同时收发消息 |

## 快速开始

### 前置条件

- Go 1.22+
- protoc（可选，用于重新生成 protobuf 代码）

### 运行

```bash
# 启动服务端
make run

# 另一个终端，运行客户端测试所有 RPC 模式
make client
```

### 构建

```bash
# 编译 server 和 client 到 bin/ 目录
make build

# 运行
./bin/server
./bin/client
```

### 测试

```bash
make test
```

### 生成 protobuf 代码

```bash
make proto
```

## 特性

- **拦截器**: 统一的日志记录和 panic 恢复
- **健康检查**: 注册 gRPC 标准健康检查服务
- **优雅关闭**: 监听 SIGINT/SIGTERM 信号，优雅停机
- **gRPC Reflection**: 支持 grpcurl 等调试工具

## Docker

```bash
docker build -t grpcdemo .
docker run -p 9000:9000 grpcdemo
```
