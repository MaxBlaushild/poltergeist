FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./admin ./admin

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/admin/cmd/server

# Build the Go application
RUN go build -o main /app/admin/cmd/server

# Expose the port on which the server will run
EXPOSE 9093

# Run the binary
CMD ["/app/main"]