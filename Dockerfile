FROM golang:1.22-alpine as builder

WORKDIR /app

ADD go.mod go.sum /app/

RUN go mod download

ADD . /app/

RUN CGO_ENABLE=0 go build -o /app/factorio-port-fixer

FROM alpine:3.20.3

RUN apk add --no-cache ca-certificates curl

EXPOSE 34197/udp 34197/tcp

COPY --from=builder /app/factorio-port-fixer /factorio-port-fixer

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD curl --fail http://localhost:34197/health || exit 1

CMD ["/factorio-port-fixer", "remote"]

