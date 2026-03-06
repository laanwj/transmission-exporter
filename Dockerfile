FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /transmission-exporter ./cmd/transmission-exporter

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /transmission-exporter /usr/bin/transmission-exporter

EXPOSE 19091

ENTRYPOINT ["/usr/bin/transmission-exporter"]
