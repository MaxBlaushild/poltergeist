FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./fount-of-erebos ./fount-of-erebos

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/fount-of-erebos/cmd/server

# Build the Go application
RUN go build -o main /app/fount-of-erebos/cmd/server

# Expose the port on which the server will run
EXPOSE 8081

# Run the binary
CMD ["/app/main"]