FROM golang:1.21 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/billing ./billing

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/billing/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/billing

FROM scratch

COPY --from=builder /usr/home/app/billing/*.env /etc/billing/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/billing /usr/local/bin/billing

WORKDIR /etc/billing

EXPOSE 8089

ENTRYPOINT ["/usr/local/bin/billing"]