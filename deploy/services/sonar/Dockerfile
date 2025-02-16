FROM golang:1.22 AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/sonar ./sonar

# RUN cd ./trivai
# RUN go mod download
WORKDIR /usr/home/app/sonar/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/sonar

FROM scratch

COPY --from=builder /usr/home/app/sonar/*.env /etc/sonar/
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/bin/sonar /usr/local/bin/sonar

WORKDIR /etc/sonar

EXPOSE 8042

ENTRYPOINT ["/usr/local/bin/sonar"]