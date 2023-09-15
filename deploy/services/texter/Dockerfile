FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./texter ./texter

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/texter/cmd/server

# Build the Go application
RUN go build -o main /app/texter/cmd/server

# Expose the port on which the server will run
EXPOSE 8084

# Run the binary
CMD ["/app/main"]