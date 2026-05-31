# 编码规范

本项目遵循 Go 官方编码规范，补充以下项目特定约定。

## 目录结构

```
internal/
├── <service>/        # 业务服务包，按服务名命名
│   ├── handler.go    # 服务实现，统一使用 Handler 结构体
│   └── handler_test.go
└── interceptor/      # gRPC 拦截器，按职责拆分文件
    ├── logging.go    # 日志拦截器
    └── recovery.go   # panic 恢复拦截器
```

## 命名规范

### 文件命名
- 服务实现：统一使用 `handler.go`
- 测试文件：`*_test.go`
- 拦截器：按职责命名，如 `logging.go`、`recovery.go`

### 结构体命名
- gRPC 服务实现：统一使用 `Handler`
- 构造函数：`NewHandler(server *grpc.Server) *Handler`

### 包级常量
- 使用 `camelCase`
- 魔法数字必须提取为常量

```go
const streamCount = 6

for n := 0; n <= streamCount; n++ {
    // ...
}
```

## Import 规范

使用 `goimports` 自动分组，顺序如下：

```go
import (
    // 1. 标准库
    "context"
    "log"

    // 2. 第三方库
    "google.golang.org/grpc"

    // 3. 内部包
    pb "gRPCServerDemo/gen/chat"
    "gRPCServerDemo/internal/chat"
)
```

### 包别名规则
- 生成代码包：使用 `pb` 前缀，如 `genchat`、`genstream`
- 当包名冲突时：使用有意义的别名，如 `genadmin`

## Proto 规范

### 文件格式
- 缩进：4 空格
- `option go_package` 以分号结尾

```protobuf
syntax = "proto3";

package proto;
option go_package = "gRPCServerDemo/chat";

message RequestMessage {
    string body = 1;
}

service ChatService {
    rpc SayHello (RequestMessage) returns (ResponseMessage) {};
}
```

### go_package 命名
- 格式：`gRPCServerDemo/<package>`
- 示例：`gRPCServerDemo/chat`、`gRPCServerDemo/stream`

## 测试规范

### Mock 结构体
- 共用方法提取为嵌入的 base mock

```go
type baseMockStream struct{}

func (m *baseMockStream) SetHeader(metadata.MD) error  { return nil }
func (m *baseMockStream) SendHeader(metadata.MD) error  { return nil }
func (m *baseMockStream) SetTrailer(metadata.MD)        {}
func (m *baseMockStream) Context() context.Context       { return context.Background() }
func (m *baseMockStream) SendMsg(interface{}) error      { return nil }
func (m *baseMockStream) RecvMsg(interface{}) error      { return nil }

type mockListStream struct {
    baseMockStream
    sent []*pb.StreamResponse
}
```

### 测试函数命名
- 格式：`Test<Struct>_<Method>`
- 子测试：`t.Run("scenario", func(t *testing.T) {...})`

```go
func TestHandler_List(t *testing.T) {
    t.Run("normal", func(t *testing.T) {...})
    t.Run("nil pt returns InvalidArgument", func(t *testing.T) {...})
}
```

## 错误处理

### gRPC 错误码
- 使用 `google.golang.org/grpc/status` 包
- 参数校验失败：`codes.InvalidArgument`
- 内部错误：`codes.Internal`

```go
if r.Pt == nil {
    return status.Error(codes.InvalidArgument, "pt is required")
}
```

### Panic 恢复
- 服务端必须使用 recovery 拦截器
- 恢复后返回 `codes.Internal` 错误

## 注释规范

### 结构体注释
```go
// Handler implements the ChatService gRPC service.
type Handler struct {
    pb.UnimplementedChatServiceServer
}
```

### 函数注释
```go
// ListServices returns the full names of all registered gRPC services on the server.
func (h *Handler) ListServices(ctx context.Context, req *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
    // ...
}
```

## 工具链

- 格式化：`goimports`
- 检查：`golangci-lint`
- 测试：`go test -race -v ./...`
