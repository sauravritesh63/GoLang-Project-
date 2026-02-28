# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree and build a statically-linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/api /api

# Environment variables with sensible defaults.
ENV PORT=8080 \
    DATABASE_URL="" \
    GIN_MODE=release

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/api"]
