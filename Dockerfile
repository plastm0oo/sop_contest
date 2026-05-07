FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app ./app

FROM alpine:3.20

WORKDIR /

COPY --from=builder /app /app

EXPOSE 8080

CMD ["/app"]