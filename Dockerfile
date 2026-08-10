# ---- Stage 1: build ----
FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 静态编译，无 cgo 依赖
RUN CGO_ENABLED=0 make build

# ---- Stage 2: runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 appuser

COPY --from=builder /src/bin/task-manager-api /app/task-manager-api
COPY --from=builder /src/configs /app/configs

WORKDIR /app

# 默认监听端口 8080，可通过环境变量覆盖（SERVER_HTTP_ADDR=0.0.0.0:8080）
ENV SERVER_HTTP_ADDR=0.0.0.0:8080
USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/task-manager-api"]
CMD ["-conf", "/app/configs"]
