# Use the official Go image
FROM golang:1.24-alpine

# Set the working directory
WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Copy UI and data files to working directory
# (or adjust paths in your Go code to find them in /app)

# Expose port 8000
EXPOSE 8000

# Command to run the application
CMD ["./main"]