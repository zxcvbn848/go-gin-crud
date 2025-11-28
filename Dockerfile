# 使用官方 Go 映像作為構建階段
FROM golang:1.25-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝必要的工具
RUN apk add --no-cache git

# 複製 go mod 文件
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# 使用輕量級的 Alpine 映像作為運行階段
FROM alpine:latest

# 安裝 CA 證書（用於 HTTPS 請求）
RUN apk --no-cache add ca-certificates tzdata

# 設置時區
ENV TZ=Asia/Taipei

WORKDIR /root/

# 從構建階段複製二進制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 運行應用
CMD ["./main"]

