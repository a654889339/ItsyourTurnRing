# 构建阶段 - 使用 Debian-based 镜像避免 musl 兼容性问题
FROM golang:1.21-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o ring-server .

# 运行阶段
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata libsqlite3-0 wget \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

RUN groupadd -g 1000 ring && \
    useradd -u 1000 -g ring -s /bin/sh -m ring

WORKDIR /app

COPY --from=builder /app/ring-server .

RUN mkdir -p /app/data /app/uploads && chown -R ring:ring /app

USER ring

EXPOSE 5102

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5102/health || exit 1

ENTRYPOINT ["./ring-server"]
