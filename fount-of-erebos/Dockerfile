# Use the official Golang image as the base image
FROM golang:1.18

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/server

# Expose the port on which the server will run
EXPOSE 8081

# Run the binary
CMD ["/app/main"]