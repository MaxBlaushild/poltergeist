# Use the official Golang image as the base image
FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./trivai ./trivai

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/trivai/cmd/server

# Build the Go application
RUN go build -o main /app/trivai/cmd/server

# Expose the port on which the server will run
EXPOSE 8087

# Run the binary
CMD ["/app/main"]