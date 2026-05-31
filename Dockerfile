FROM --platform=$BUILDPLATFORM golang:1.22 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# 设置中国代理
ENV GOPROXY=https://goproxy.cn,direct

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码
COPY cmd/ cmd/
COPY internal/ internal/
COPY gen/ gen/

# 编译
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o server ./cmd/server

# 最小镜像
FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /workspace/server .
USER nonroot:nonroot

EXPOSE 9000
ENTRYPOINT ["/server"]
