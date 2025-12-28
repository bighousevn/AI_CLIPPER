FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./ 
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/worker ./cmd/worker/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/worker-email ./cmd/worker-email/main.go

FROM alpine:latest as ai-clipper-be

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN mkdir -p /app/cmd/server \
             /app/cmd/worker \
             /app/cmd/worker-email \
             /app/database/migrations

COPY --from=builder /app/database/migrations /app/database/migrations

COPY --from=builder /app/bin/server /app/cmd/server/server
COPY --from=builder /app/bin/worker /app/cmd/worker/worker
COPY --from=builder /app/bin/worker-email /app/cmd/worker-email/worker-email

COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

EXPOSE 8080

CMD ["/app/start.sh"]