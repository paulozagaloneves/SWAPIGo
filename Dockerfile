# syntax=docker/dockerfile:1

# -------- Builder stage --------
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Install CA certificates (useful for HTTPS requests in build/test)
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Cache deps
COPY go.mod go.sum ./
# Align module Go version for toolchain compatibility inside container
RUN go mod edit -go=1.22 && go mod download

# Copy the rest
COPY . .

RUN go mod edit -go=1.22

# Build static binary
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /out/swapigo ./

# -------- Runtime stage --------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -H -u 10001 appuser
USER appuser
WORKDIR /home/appuser

COPY --from=builder /out/swapigo /usr/local/bin/swapigo

ENV PORT=8080
EXPOSE 8080

CMD ["/usr/local/bin/swapigo"]
