# Stage 1: Build the application
   
    FROM golang:1.22-alpine AS builder
    
   
    WORKDIR /app
    
  
    COPY go.mod go.sum ./

    RUN go mod download
    
  
    COPY . .
    
    
    RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o /main ./cmd/server/main.go
    
# Stage 2: Tạo image cuối cùng, tối giản
    FROM alpine:latest
    
    WORKDIR /root/
    
    COPY --from=builder /main .
    
   
    
    EXPOSE 8080
    
    CMD ["./main"]