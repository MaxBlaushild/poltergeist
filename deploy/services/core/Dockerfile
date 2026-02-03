# syntax=docker/dockerfile:1.7
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/core/go.mod ./go/core/go.sum ./core/
COPY ./go/pkg ./pkg
COPY ./go/sonar/go.mod ./go/sonar/go.sum ./sonar/
COPY ./go/travel-angels/go.mod ./go/travel-angels/go.sum ./travel-angels/
COPY ./go/verifiable-sn/go.mod ./go/verifiable-sn/go.sum ./verifiable-sn/

WORKDIR /usr/home/app/core
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/core ./core
# COPY ./go/final-fete ./final-fete
COPY ./go/sonar ./sonar
COPY ./go/travel-angels ./travel-angels
COPY ./go/verifiable-sn ./verifiable-sn

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/core/cmd/server

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/core

FROM scratch

COPY --from=builder /usr/home/app/core/*.env /etc/core/
COPY --from=builder /usr/home/app/verifiable-sn/vera-firebase.json /etc/core/vera-firebase.json
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/core /usr/local/bin/core

WORKDIR /etc/core

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/core"]
