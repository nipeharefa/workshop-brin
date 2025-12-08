# syntax=docker/dockerfile:1

# Build stage
ARG TARGETPLATFORM
ARG BUILDPLATFORM

FROM --platform=$BUILDPLATFORM golang:1.25-trixie AS builder

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
# 
COPY . .

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags="-w -s" -trimpath -o build/go-service ./cmd/bot/main.go

# Runtime stage
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/build/go-service /app/go-service
COPY migrations migrations
# Set non-root user
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["./go-service"]
