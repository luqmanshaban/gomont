# --- Build stage ---
FROM golang:1.26.3-alpine AS builder

WORKDIR /app

# Cache module downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO disabled for a fully static binary that runs in a scratch-like image
RUN CGO_ENABLED=0 GOOS=linux go build -o gomont ./main.go

# --- Runtime stage ---
FROM alpine:latest

# ca-certificates needed for outbound HTTPS (monitor pings, SMTP over TLS)
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/gomont .

EXPOSE 8080

CMD ["./gomont"]