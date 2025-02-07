FROM golang:1.21-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy source code first so we can initialize the module properly
COPY . .

# Initialize and tidy the module
RUN go mod download && \
    go mod tidy

# Build the application
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]