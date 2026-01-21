FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/ethereum-transactor ./ethereum-transactor

WORKDIR /usr/home/app/ethereum-transactor/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/ethereum-transactor

FROM scratch

COPY --from=builder /usr/home/app/ethereum-transactor/*.env /etc/ethereum-transactor/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/ethereum-transactor /usr/local/bin/ethereum-transactor

WORKDIR /etc/ethereum-transactor

EXPOSE 8088

ENTRYPOINT ["/usr/local/bin/ethereum-transactor"]

