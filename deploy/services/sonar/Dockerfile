# syntax=docker/dockerfile:1.7
FROM golang:1.23 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/sonar/go.mod ./go/sonar/go.sum ./sonar/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/sonar
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/sonar ./sonar

WORKDIR /usr/home/app/sonar/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/sonar


FROM scratch

COPY --from=builder /usr/home/app/sonar/*.env /etc/sonar/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/sonar /usr/local/bin/sonar

WORKDIR /etc/sonar

EXPOSE 8042

ENTRYPOINT ["/usr/local/bin/sonar"]