# syntax=docker/dockerfile:1.7
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/job-runner/go.mod ./go/job-runner/go.sum ./job-runner/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/job-runner
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/job-runner ./job-runner

WORKDIR /usr/home/app/job-runner/cmd/runner

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/job-runner


FROM scratch

COPY --from=builder /usr/home/app/job-runner/*.env /etc/job-runner/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/job-runner /usr/local/bin/job-runner

WORKDIR /etc/job-runner

EXPOSE 9013

ENTRYPOINT ["/usr/local/bin/job-runner"]