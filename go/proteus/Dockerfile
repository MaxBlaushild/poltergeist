# syntax=docker/dockerfile:1.7
# Use the official Golang image as the base image
FROM golang:1.18

ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o main ./cmd/server

# Expose the port on which the server will run
EXPOSE 8085

# Run the binary
CMD ["/app/main"]
