# syntax=docker/dockerfile:1.7
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/final-fete/go.mod ./go/final-fete/go.sum ./final-fete/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/final-fete
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/final-fete ./final-fete

WORKDIR /usr/home/app/final-fete/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/final-fete


FROM scratch

COPY --from=builder /usr/home/app/final-fete/*.env /etc/final-fete/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/final-fete /usr/local/bin/final-fete

WORKDIR /etc/final-fete

EXPOSE 8085

ENTRYPOINT ["/usr/local/bin/final-fete"]