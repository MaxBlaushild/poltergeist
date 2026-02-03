# syntax=docker/dockerfile:1.7
FROM golang:1.21 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/billing/go.mod ./go/billing/go.sum ./billing/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/billing
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/billing ./billing

WORKDIR /usr/home/app/billing/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/billing


FROM scratch

COPY --from=builder /usr/home/app/billing/*.env /etc/billing/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/billing /usr/local/bin/billing

WORKDIR /etc/billing

EXPOSE 8089

ENTRYPOINT ["/usr/local/bin/billing"]