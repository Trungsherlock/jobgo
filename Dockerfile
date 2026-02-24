# --- Build stage ---
FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o jobgo ./cmd/jobgo

# --- Runtime stage ---
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/jobgo .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/data ./data

EXPOSE 8080

CMD ["./jobgo", "serve"]