# ============================================================
# BUILD STAGE
# ============================================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Required build tools
RUN apk add --no-cache git ca-certificates

# Copy dependency files first for Docker layer caching
COPY go.mod go.sum ./

RUN go mod download

# Copy application source
COPY . .

# Build REST API
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o nid-backend ./cmd/main.go


# ============================================================
# RUNTIME STAGE
# ============================================================
FROM alpine:3.22

WORKDIR /app

# HTTPS/TLS certificates
RUN apk add --no-cache ca-certificates

# Copy production binary
COPY --from=builder /app/nid-backend .

# REST API
EXPOSE 8081

# Start API
CMD ["./nid-backend"]
