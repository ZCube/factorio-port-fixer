FROM golang:1.19 as builder

WORKDIR /app

ADD go.mod go.sum /app/

RUN go mod download

ADD . /app/

RUN CGO_ENABLE=0 go build -o /app/factorio-port-fixer

FROM gcr.io/distroless/base-debian11

EXPOSE 34197/udp

COPY --from=builder /app/factorio-port-fixer /factorio-port-fixer

CMD ["/factorio-port-fixer"]
