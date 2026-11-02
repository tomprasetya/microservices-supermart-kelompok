# ==========================================
# STAGE 1: Build Environment (Kompilasi)
# ==========================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src

# Optimasi Layer Caching: Salin manifes dependensi terlebih dahulu
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Salin source code
COPY . .

# Kompilasi statis murni tanpa libc dinamik (-ldflags="-s -w" memangkas tabel simbol DWARF)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/order-service ./cmd/publisher/main.go

# Buat pengguna non-privilese UID 10001
RUN adduser -D -g '' -u 10001 appuser

# ==========================================
# STAGE 2 - VARIAN A: Alpine (musl libc minimal)
# ==========================================
FROM alpine:3.19 AS runtime-alpine
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/order-service /app/order-service
COPY --from=builder /etc/passwd /etc/passwd
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/order-service"]

# ==========================================
# STAGE 2 - VARIAN B: Google Distroless (Tanpa Package Manager / Shell)
# ==========================================
FROM gcr.io/distroless/static-debian12:nonroot AS runtime-distroless
WORKDIR /app
COPY --from=builder /bin/order-service /app/order-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/order-service"]

# ==========================================
# STAGE 2 - VARIAN C: Scratch (0-byte Base Image)
# ==========================================
FROM scratch AS runtime-scratch
WORKDIR /app
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /bin/order-service /app/order-service
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/order-service"]
