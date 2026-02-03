# syntax=docker/dockerfile:1.7
FROM golang:1.21 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/trivai/go.mod ./go/trivai/go.sum ./trivai/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/trivai
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/trivai ./trivai

WORKDIR /usr/home/app/trivai/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/trivai


FROM scratch

COPY --from=builder /usr/home/app/trivai/*.env /etc/trivai/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/trivai /usr/local/bin/trivai

WORKDIR /etc/trivai

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/trivai"]