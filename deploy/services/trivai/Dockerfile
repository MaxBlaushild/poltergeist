FROM golang:1.21 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/trivai ./trivai

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/trivai/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/trivai

FROM scratch

COPY --from=builder /usr/home/app/trivai/*.env /etc/trivai/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/trivai /usr/local/bin/trivai

WORKDIR /etc/trivai

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/trivai"]