# syntax=docker/dockerfile:1
# ========== 阶段1：编译 ==========
FROM golang:1.24-alpine AS builder

# 使用国内 Go 代理加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build

# 复制 go.mod 和 go.sum，利用层缓存
COPY go.mod go.sum ./

# 复制 plugin-sdk（本地依赖，go.mod replace 指向 ../plugin-sdk，即 /plugin-sdk）
COPY plugin-sdk /plugin-sdk

# 下载依赖（挂载 Go module 缓存，跨构建复用）
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# ========== 阶段2：运行 ==========
FROM alpine:3.20

# 使用国内 Alpine 镜像源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/main /app/main

# 根据环境选择配置文件
ARG ENV=test
COPY --from=builder /build/config/settings-${ENV}.yml /app/config/settings.yml

# 创建日志目录并设置权限
RUN chmod +x /app/main && mkdir -p /app/temp/logs && chown -R 1001:1001 /app

EXPOSE 8000

# 以非 root 用户运行（生产安全最佳实践）
USER 1001

ENTRYPOINT ["/app/main"]
CMD ["server", "-c", "/app/config/settings.yml"]
