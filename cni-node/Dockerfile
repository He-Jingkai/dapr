FROM golang:1.19 AS builder
WORKDIR /src
COPY . .
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o /go/bin/cni-node .

FROM gcr.io/istio-release/base:latest
COPY --from=builder --chown=root:root /go/bin/cni-node /bin/cni-node
CMD ["/bin/cni-node"]