FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./authenticator ./authenticator

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/authenticator/cmd/server

# Build the Go application
RUN go build -o main /app/authenticator/cmd/server

# Expose the port on which the server will run
EXPOSE 8089

# Run the binary
CMD ["/app/main"]