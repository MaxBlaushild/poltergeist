FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./crystal-crisis-api ./crystal-crisis-api

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/crystal-crisis-api/cmd/server

# Build the Go application
RUN go build -o main /app/crystal-crisis-api/cmd/server

# Expose the port on which the server will run
EXPOSE 8091

# Run the binary
CMD ["/app/main"]