FROM golang:1.18 AS builder


ENV CGO_ENABLED=0
ENV GOPATH=/usr/home/app

# Set the working directory
WORKDIR /usr/home/app

COPY ./go/pkg ./pkg
COPY ./go/admin ./admin

WORKDIR /usr/home/app/admin/cmd/server

# Build the Go application
RUN go build -installsuffix 'static' -o /usr/local/bin/admin

FROM postgres:latest

RUN echo "host poltergeist db_user 10.0.34.124/32 md5" >> "$PGDATA/pg_hba.conf"
COPY --from=builder /usr/home/app/admin/*.env /etc/admin/
COPY --from=builder /usr/home/app/admin/internal/migrate/admin.pgsql /etc/admin/internal/migrate/admin.pgsql
COPY --from=builder /usr/home/app/admin/internal/migrate/admin.pgsql /usr/local/bin/internal/migrate/admin.pgsql
COPY --from=builder /usr/local/bin/admin /usr/local/bin/admin
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

WORKDIR /etc/admin

EXPOSE 9093

ENTRYPOINT ["/usr/local/bin/admin"]