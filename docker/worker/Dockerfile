# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/worker ./cmd/worker

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/worker /worker

# Environment variables with sensible defaults.
ENV WORKER_ID=worker-1 \
    LOG_LEVEL=info

USER nonroot:nonroot
ENTRYPOINT ["/worker"]
