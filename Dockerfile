# ---- Stage 1: build ----
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /app/task-manager-api ./cmd/task-manager-api

# ---- Stage 2: runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 appuser
COPY --from=builder /app/task-manager-api /app/task-manager-api
COPY --from=builder /src/configs /app/configs
WORKDIR /app
ENV SERVER_HTTP_ADDR=0.0.0.0:8080
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/task-manager-api"]
CMD ["-conf", "/app/configs"]
