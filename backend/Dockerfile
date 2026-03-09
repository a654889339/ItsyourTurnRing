FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache gcc musl-dev

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装必要的库
RUN apk add --no-cache ca-certificates tzdata sqlite-libs wget

# 复制编译好的二进制
COPY --from=builder /app/main .

# 创建数据和上传目录
RUN mkdir -p data uploads

EXPOSE 5102

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5102/health || exit 1

CMD ["./main"]
