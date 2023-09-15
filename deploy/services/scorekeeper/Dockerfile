FROM golang:1.18

ENV GOPATH=/app

# Set the working directory
WORKDIR /app

COPY ./pkg ./pkg
COPY ./scorekeeper ./scorekeeper

# RUN cd ./trivai
# RUN go mod download
WORKDIR /app/scorekeeper/cmd/server

# Build the Go application
RUN go build -o main /app/scorekeeper/cmd/server

# Expose the port on which the server will run
EXPOSE 8086

# Run the binary
CMD ["/app/main"]