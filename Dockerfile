# ─── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Install git (needed for go modules with VCS dependencies)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache module downloads as a separate layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a fully static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/server ./cmd/server/main.go

# ─── Stage 2: Runtime ────────────────────────────────────────────────────────
FROM scratch

# Copy timezone data and CA certs from builder (needed for TLS + time zones)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /bin/server /server

# Railway and most platforms inject PORT at runtime — default 8080
EXPOSE 8080

ENTRYPOINT ["/server"]
