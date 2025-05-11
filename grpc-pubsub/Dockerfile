# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /server ./cmd/server

# Runtime stage
FROM alpine:latest
COPY --from=builder /server /server
COPY config.yaml /config.yaml
EXPOSE 50051
ENTRYPOINT ["/server"]
