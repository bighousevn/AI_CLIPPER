# --- Stage 1: Builder ---
FROM golang:alpine AS builder

# Cài đặt git
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency
COPY go.mod go.sum ./ 
RUN go mod download

# Copy source
COPY . .

# Build 3 binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/worker ./cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/worker-email ./cmd/worker-email/main.go

# --- Stage 2: Runtime ---
FROM alpine:latest

# Cài thư viện hệ thống
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Tạo cấu trúc thư mục giống source gốc để đảm bảo logic relative path (../../) hoạt động đúng
RUN mkdir -p /app/cmd/server \
             /app/cmd/worker \
             /app/cmd/worker-email \
             /app/database/migrations

# Copy migrations
COPY --from=builder /app/database/migrations /app/database/migrations

# Copy binaries vào đúng thư mục cấu trúc gốc
COPY --from=builder /app/bin/server /app/cmd/server/server
COPY --from=builder /app/bin/worker /app/cmd/worker/worker
COPY --from=builder /app/bin/worker-email /app/cmd/worker-email/worker-email

# Expose port
EXPOSE 8080

# Mặc định chạy Server. 
# Khi muốn chạy Worker, chỉ cần override CMD trong docker-compose hoặc lệnh docker run
# Ví dụ: docker run ... /app/cmd/worker/worker
WORKDIR /app/cmd/server
CMD ["./server"]
