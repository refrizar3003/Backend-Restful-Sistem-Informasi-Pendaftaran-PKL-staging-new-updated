# Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags="-s -w" .

# Final
FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]