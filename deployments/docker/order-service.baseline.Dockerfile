FROM golang:1.23

WORKDIR /app

# Anti-pattern: langsung salin seluruh source tree
COPY . .

# Kompilasi standar tanpa strip simbol debug dan CGO aktif
RUN go build -o /app/order-service ./cmd/publisher/main.go

EXPOSE 8080

CMD ["/app/order-service"]
