FROM golang:1.24.2

WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the app
RUN go build -o main .

# Run the binary
CMD ["./main"]
