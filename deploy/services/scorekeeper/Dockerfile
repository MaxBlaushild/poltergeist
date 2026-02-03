# syntax=docker/dockerfile:1.7
FROM golang:1.18 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/scorekeeper/go.mod ./go/scorekeeper/go.sum ./scorekeeper/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/scorekeeper
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/scorekeeper ./scorekeeper

WORKDIR /usr/home/app/scorekeeper/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/scorekeeper


FROM scratch

COPY --from=builder /usr/home/app/scorekeeper/*.env /etc/scorekeeper/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/scorekeeper /usr/local/bin/scorekeeper

WORKDIR /etc/scorekeeper

EXPOSE 8086

ENTRYPOINT ["/usr/local/bin/scorekeeper"]