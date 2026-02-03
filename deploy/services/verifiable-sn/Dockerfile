# syntax=docker/dockerfile:1.7
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/verifiable-sn/go.mod ./go/verifiable-sn/go.sum ./verifiable-sn/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/verifiable-sn
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/verifiable-sn ./verifiable-sn

WORKDIR /usr/home/app/verifiable-sn/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/verifiable-sn


FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/verifiable-sn /usr/local/bin/verifiable-sn

EXPOSE 8087

ENTRYPOINT ["/usr/local/bin/verifiable-sn"]

