# syntax=docker/dockerfile:1.7
#
# reef-site's HTTP server only talks to Postgres and enqueues jobs — it never
# shells out to OpenSCAD/a slicer directly (R-2.5's subprocess sandbox lives in
# go/job-runner, which actually executes generation/slice jobs; see
# deploy/services/job-runner/Dockerfile). So this stays on the same `scratch`
# pattern every other service in this repo uses, with nothing extra needed.
#
# This Dockerfile matters only if reef-site is later split out of go/core into
# its own deployed ECS service (see go/reef-site/INVENTORY.md) — today the
# storefront is served by the composed core image, same as every other domain.
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

WORKDIR /usr/home/app

COPY ./go/reef-site/go.mod ./go/reef-site/go.sum ./reef-site/
COPY ./go/pkg ./pkg

WORKDIR /usr/home/app/reef-site
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/reef-site ./reef-site

WORKDIR /usr/home/app/reef-site/cmd/server

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -installsuffix 'static' -o /usr/local/bin/reef-site


FROM scratch

COPY --from=builder /usr/home/app/reef-site/*.env /etc/reef-site/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/reef-site /usr/local/bin/reef-site

WORKDIR /etc/reef-site

EXPOSE 8091

ENTRYPOINT ["/usr/local/bin/reef-site"]
