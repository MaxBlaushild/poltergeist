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
FROM debian:bookworm-slim

ARG PRUSASLICER_VERSION=2.7.4
ARG PRUSASLICER_URL=https://github.com/prusa3d/PrusaSlicer/releases/download/version_${PRUSASLICER_VERSION}/PrusaSlicer-${PRUSASLICER_VERSION}+linux-x64-GTK3-202404050928.AppImage

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      openscad \
      ca-certificates \
      wget \
      libgtk-3-0 \
      libglu1-mesa \
      libgl1 \
      libegl1 \
      libwayland-client0 \
      libwayland-egl1 \
      libdbus-1-3 \
      liblzma5 \
      libfontconfig1 \
    && rm -rf /var/lib/apt/lists/*

# Extract the PrusaSlicer AppImage's embedded squashfs directly instead of
# running the AppImage's own self-extraction (`--appimage-extract`) — that
# stub reliably fails with "Exec format error" when this image is
# cross-built for linux/amd64 under QEMU emulation (e.g. from Apple Silicon;
# see `make job-runner/build`), even though plain ELF binaries emulate fine.
# The squashfs magic ("hsqs") can appear more than once as an incidental
# byte match earlier in the file, so every candidate offset is tried until
# one actually unpacks.
RUN apt-get update && \
    apt-get install -y --no-install-recommends squashfs-tools python3 && \
    rm -rf /var/lib/apt/lists/* && \
    wget -q -O /tmp/prusaslicer.AppImage "$PRUSASLICER_URL" && \
    python3 -c "import re; data = open('/tmp/prusaslicer.AppImage', 'rb').read(); [print(m.start()) for m in re.finditer(b'hsqs', data)]" > /tmp/prusaslicer-offsets.txt && \
    cd /opt && \
    while read -r offset; do \
      rm -rf /opt/squashfs-root; \
      unsquashfs -o "$offset" -d /opt/squashfs-root /tmp/prusaslicer.AppImage > /dev/null 2>&1 && break; \
    done < /tmp/prusaslicer-offsets.txt && \
    test -x /opt/squashfs-root/AppRun && \
    mv /opt/squashfs-root /opt/prusaslicer && \
    ln -s /opt/prusaslicer/AppRun /usr/local/bin/prusa-slicer && \
    rm /tmp/prusaslicer.AppImage /tmp/prusaslicer-offsets.txt && \
    apt-get purge -y --auto-remove squashfs-tools python3

ENV REEF_OPENSCAD_BIN=/usr/bin/openscad
ENV REEF_SLICER_BIN=/usr/local/bin/prusa-slicer

COPY --from=builder /usr/home/app/job-runner/*.env /etc/job-runner/
COPY --from=builder /usr/local/bin/job-runner /usr/local/bin/job-runner

WORKDIR /etc/job-runner

EXPOSE 9013

ENTRYPOINT ["/usr/local/bin/job-runner"]
