FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app ./cmd/server

FROM alpine:3.23

WORKDIR /

COPY --from=builder /app /app
COPY --from=builder /src/migrations /migrations

EXPOSE 8080

CMD ["/app"]