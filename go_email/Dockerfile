# Stage 1: Build the Go binary
FROM golang:1.24.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o go_email .

# Stage 2: Run the binary
FROM alpine:3.20

# Install timezone data and ca-certificates for HTTPS/MongoDB
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/go_email .
COPY --from=builder /app/mail/templates ./mail/templates

# Set timezone environment variable
ENV TZ=Asia/Kolkata

# Env vars will be injected at runtime

EXPOSE 8080
CMD ["./go_email"]
