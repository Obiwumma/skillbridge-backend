# BUILD STAGE
FROM golang:1.26-alpine AS builder

# Install system dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Compile application with optimized static flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /skillbridge-api ./cmd/api

# FINAL STAGE
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy static binary from builder
COPY --from=builder /skillbridge-api .

# Copy DDL migrations so the system can run startup checks
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./skillbridge-api"]
