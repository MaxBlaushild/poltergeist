FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/core ./core
# COPY ./go/final-fete ./final-fete
COPY ./go/sonar ./sonar
COPY ./go/travel-angels ./travel-angels
COPY ./go/verifiable-sn ./verifiable-sn

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/core/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/core

FROM scratch

COPY --from=builder /usr/home/app/core/*.env /etc/core/
COPY --from=builder /usr/home/app/verifiable-sn/vera-firebase.json /etc/core/vera-firebase.json
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/core /usr/local/bin/core

WORKDIR /etc/core

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/core"]