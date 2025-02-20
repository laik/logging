# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
ENV GO111MODULE=on
ENV CGO_ENABLED=0 GOPROXY=https://goproxy.cn,https://goproxy.io,https://mirrors.aliyun.com/goproxy/,https://athens.azurefd.net,direct

COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Add the go source
ADD . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --ldflags "-extldflags -static"  -o build/api-server cmd/api-server/*.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/build/api-server .

ENTRYPOINT ["./api-server"]