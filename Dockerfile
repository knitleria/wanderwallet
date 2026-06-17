FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /wanderwallet ./cmd/wanderwallet

FROM alpine:3.20

RUN adduser -D appuser

WORKDIR /app
COPY --from=builder /wanderwallet /app/wanderwallet

USER appuser

ENV RUN_ADDRESS=:8080
EXPOSE 8080

CMD ["/app/wanderwallet"]
