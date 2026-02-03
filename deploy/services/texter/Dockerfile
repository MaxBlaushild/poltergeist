# syntax=docker/dockerfile:1.7
FROM golang:1.18 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/texter/go.mod ./go/texter/go.sum ./texter/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/texter
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/texter ./texter

WORKDIR /usr/home/app/texter/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/texter


FROM scratch

COPY --from=builder /usr/home/app/texter/*.env /etc/texter/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/texter /usr/local/bin/texter

WORKDIR /etc/texter

EXPOSE 8084

ENTRYPOINT ["/usr/local/bin/texter"]