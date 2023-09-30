FROM golang:1.18 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/fount-of-erebos ./fount-of-erebos

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/fount-of-erebos/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/fount-of-erebos

FROM scratch

COPY --from=builder /usr/home/app/fount-of-erebos/*.env /etc/fount-of-erebos/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/fount-of-erebos /usr/local/bin/fount-of-erebos

WORKDIR /etc/fount-of-erebos

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/fount-of-erebos"]