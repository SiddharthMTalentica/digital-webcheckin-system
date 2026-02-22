FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/init_db.sh .
RUN chmod +x init_db.sh

# Expose API port
EXPOSE 8080

CMD ["./main"]
