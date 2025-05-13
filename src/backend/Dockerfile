FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./
RUN go run . -scrapeonly
RUN echo "Isi direktori /app setelah scrapeonly:" && ls -la /app
RUN echo "Isi direktori /app/data setelah scrapeonly:" && ls -la /app/data || echo "/app/data tidak ditemukan atau kosong"

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/main_backend .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/data ./data/

COPY --from=builder /app/main_backend .
EXPOSE 8080
CMD ["./main_backend"]