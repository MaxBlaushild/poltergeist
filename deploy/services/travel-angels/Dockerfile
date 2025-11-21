FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/travel-angels ./travel-angels

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/travel-angels/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/travel-angels

FROM scratch

COPY --from=builder /usr/home/app/travel-angels/*.env /etc/travel-angels/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/travel-angels /usr/local/bin/travel-angels

WORKDIR /etc/travel-angels

EXPOSE 8083

ENTRYPOINT ["/usr/local/bin/travel-angels"]