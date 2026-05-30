.PHONY: proto build run client test lint clean

# 生成 protobuf 代码
proto:
	protoc --go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		proto/chat.proto
	protoc --go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		proto/stream.proto

# 编译 server 和 client
build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# 运行 server
run:
	go run ./cmd/server

# 运行 client
client:
	go run ./cmd/client

# 运行测试（含 race 检测）
test:
	go test -race -v ./...

# 代码检查
lint:
	golangci-lint run ./...

# 清理编译产物
clean:
	rm -rf bin/
