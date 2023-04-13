FROM golang:alpine AS builder

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GIN_MODE=release \
    PORT=80 \
    CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /build

COPY . ./
#build出二进制文件app
RUN go build -o app .

FROM alpine:latest
#第二段打包
COPY --from=builder /build/app /
#项目端口
EXPOSE 8080
#以二进制方式执行
ENTRYPOINT ["/app"]