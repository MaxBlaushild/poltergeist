FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/final-fete ./final-fete

# RUN cd ./final-fete
# RUN go mod download
WORKDIR /usr/home/app/final-fete/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/final-fete

FROM scratch

COPY --from=builder /usr/home/app/final-fete/*.env /etc/final-fete/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/final-fete /usr/local/bin/final-fete

WORKDIR /etc/final-fete

EXPOSE 8085

ENTRYPOINT ["/usr/local/bin/final-fete"]