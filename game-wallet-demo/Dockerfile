# STAGE 1: The Builder (Heavy, contains tools)
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git openssl

WORKDIR /app

# Cache Dependencies (Layer Caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy Source Code
COPY . .

# Generate Prisma Client
RUN go run github.com/steebchen/prisma-client-go generate

# Build the Binary
# -ldflags="-s -w": Strips debug information (Reduces binary size by ~25%)
# CGO_ENABLED=0: Ensures a static binary that works on empty Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# --------------------------------------------------------------------
# STAGE 2: The Runner (Tiny, Production Only)
FROM alpine:latest

# openssl: Required by Prisma Query Engine
RUN apk add --no-cache ca-certificates openssl

WORKDIR /app

COPY --from=builder /app/main .

# Copying the Prisma Engine so the app can find the query engine.
COPY --from=builder /app/prisma ./prisma

# Security & Cleanup
# Create a non-root user (Security Best Practice)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser


EXPOSE 8080
ENTRYPOINT ["./main"]