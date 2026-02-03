# syntax=docker/dockerfile:1.7
FROM golang:1.22 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/fount-of-erebos/go.mod ./go/fount-of-erebos/go.sum ./fount-of-erebos/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/fount-of-erebos
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/fount-of-erebos ./fount-of-erebos

WORKDIR /usr/home/app/fount-of-erebos/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/fount-of-erebos


FROM scratch

COPY --from=builder /usr/home/app/fount-of-erebos/*.env /etc/fount-of-erebos/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/fount-of-erebos /usr/local/bin/fount-of-erebos

WORKDIR /etc/fount-of-erebos

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/fount-of-erebos"]