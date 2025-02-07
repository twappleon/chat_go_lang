# 使用 golang 1.21 作為基礎鏡像
FROM golang:1.21-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝基本依賴
RUN apk add --no-cache git

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 編譯應用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用輕量級的 alpine 作為運行環境
FROM alpine:latest

WORKDIR /app

# 從 builder 階段複製編譯好的執行檔
COPY --from=builder /app/main .

# 設置時區
RUN apk add --no-cache tzdata
ENV TZ=Asia/Taipei

# 暴露端口
EXPOSE 8888

# 運行應用
CMD ["./main"]