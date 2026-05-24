FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

ADD https://github.com/pressly/goose/releases/download/v3.19.2/goose_linux_x86_64 /bin/goose
RUN chmod +x /bin/goose

EXPOSE 8080

CMD ["./api"]
