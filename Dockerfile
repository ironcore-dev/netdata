# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM alpine:edge
ARG USER=nonroot
RUN apk -U upgrade && apk add --no-cache \
    nmap \
    libcap \
    sudo \
    nmap-scripts && \
    adduser -g $USER -D $USER \
        && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
        && chmod 0440 /etc/sudoers.d/$USER \
        && setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip  /usr/bin/nmap \
    && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=builder /workspace/manager .
USER $USER:$USER
ENTRYPOINT ["/manager"]
