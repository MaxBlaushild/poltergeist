FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/verifiable-sn ./verifiable-sn

WORKDIR /usr/home/app/verifiable-sn/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/verifiable-sn

FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/verifiable-sn /usr/local/bin/verifiable-sn

EXPOSE 8087

ENTRYPOINT ["/usr/local/bin/verifiable-sn"]

