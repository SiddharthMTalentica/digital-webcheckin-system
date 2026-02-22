# Web Check-In Service

FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o webcheckin cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/webcheckin .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8081

# Run the binary
CMD ["./webcheckin"]
