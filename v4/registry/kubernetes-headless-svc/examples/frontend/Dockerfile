FROM golang:alpine AS builder

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    GIN_MODE=release \
    PORT=80 \
    CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /build

COPY . ./
RUN go build -o app .

FROM alpine:latest
COPY --from=builder /build/app /
EXPOSE 8080
ENTRYPOINT ["/app"]