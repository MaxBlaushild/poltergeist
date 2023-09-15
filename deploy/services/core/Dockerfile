FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./core ./core

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/core/cmd/server

# Build the Go application
RUN go build -o main /app/core/cmd/server

# Expose the port on which the server will run
EXPOSE 8080

# Run the binary
CMD ["/app/main"]