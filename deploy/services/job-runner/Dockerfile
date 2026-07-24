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


# job-runner used to ship on `scratch` (every existing task type is pure
# Go/HTTP/DB work). reef-site's generation and slicing jobs (R-2.4/R-2.5/R-2.7)
# are the first task types that need to shell out to external binaries, so
# this image now carries OpenSCAD and a pinned PrusaSlicer build. Both
# versions are pinned here and recorded on every generated artifact per R-2.5.
#
# NOTE: not exercised with `docker build` in the environment that authored
# this — verify the PrusaSlicer AppImage URL/checksum on a real build before
# relying on it.
FROM debian:bookworm-slim

ARG PRUSASLICER_VERSION=2.7.4
ARG PRUSASLICER_URL=https://github.com/prusa3d/PrusaSlicer/releases/download/version_${PRUSASLICER_VERSION}/PrusaSlicer-${PRUSASLICER_VERSION}+linux-x64-GTK3-202312191501.AppImage

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      openscad \
      ca-certificates \
      wget \
      libfuse2 \
    && rm -rf /var/lib/apt/lists/*

# Extract the PrusaSlicer AppImage so it runs directly (no FUSE mount needed
# inside a container) at a fixed, version-pinned path.
RUN wget -q -O /tmp/prusaslicer.AppImage "$PRUSASLICER_URL" && \
    chmod +x /tmp/prusaslicer.AppImage && \
    cd /opt && /tmp/prusaslicer.AppImage --appimage-extract > /dev/null && \
    mv squashfs-root prusaslicer && \
    ln -s /opt/prusaslicer/AppRun /usr/local/bin/prusa-slicer && \
    rm /tmp/prusaslicer.AppImage

ENV REEF_OPENSCAD_BIN=/usr/bin/openscad
ENV REEF_SLICER_BIN=/usr/local/bin/prusa-slicer

COPY --from=builder /usr/home/app/job-runner/*.env /etc/job-runner/
COPY --from=builder /usr/local/bin/job-runner /usr/local/bin/job-runner

WORKDIR /etc/job-runner

EXPOSE 9013

ENTRYPOINT ["/usr/local/bin/job-runner"]
