# syntax=docker/dockerfile:1.7
FROM golang:1.22 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/authenticator/go.mod ./go/authenticator/go.sum ./authenticator/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/authenticator
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/authenticator ./authenticator

WORKDIR /usr/home/app/authenticator/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/authenticator


FROM scratch

COPY --from=builder /usr/home/app/authenticator/*.env /etc/authenticator/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/authenticator /usr/local/bin/authenticator

WORKDIR /etc/authenticator

EXPOSE 8089

ENTRYPOINT ["/usr/local/bin/authenticator"]