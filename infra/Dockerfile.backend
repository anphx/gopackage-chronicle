FROM golang:1.27.4-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

# Migrations are embedded in the binary via //go:embed, so nothing else is copied.
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.22

# Needed for TLS to Supabase (POSTGRES_SSLMODE defaults to require).
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /server .

EXPOSE 8080

CMD ["./server"]
