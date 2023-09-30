FROM golang:1.18 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/authenticator ./authenticator

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/authenticator/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/authenticator

FROM scratch

COPY --from=builder /usr/home/app/authenticator/*.env /etc/authenticator/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/authenticator /usr/local/bin/authenticator

WORKDIR /etc/authenticator

EXPOSE 8089

ENTRYPOINT ["/usr/local/bin/authenticator"]