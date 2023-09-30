FROM golang:1.18 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/crystal-crisis-api ./crystal-crisis-api

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/crystal-crisis-api/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/crystal-crisis-api

FROM scratch

COPY --from=builder /usr/home/app/crystal-crisis-api/live.env /etc/crystal-crisis-api/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/crystal-crisis-api /usr/local/bin/crystal-crisis-api

WORKDIR /etc/crystal-crisis-api

EXPOSE 8091

ENTRYPOINT ["/usr/local/bin/crystal-crisis-api"]