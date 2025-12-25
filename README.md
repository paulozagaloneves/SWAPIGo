# SWAPIGo REST API Microservice

A minimal Go microservice that exposes:

- GET `/healthz` — service health
- GET `/api/swapi/<resource>` — read-only proxy to https://swapi.info/api/
- GET `/` — service info

## Build and Run

Requires Go 1.20+ (recommended 1.22).

```bash
# From the project root
go run .
# or
go build -o swapigo
./swapigo
```

The server listens on port 8080 by default. Set `PORT` to override.

```bash
PORT=9090 go run .
```

## Example Requests

```bash
# Health
curl http://localhost:8080/healthz

# Service info
curl http://localhost:8080/

# SWAPI proxy examples (path-based)
curl "http://localhost:8080/api/swapi/"
curl "http://localhost:8080/api/swapi/people/1"
curl "http://localhost:8080/api/swapi/films/1"
```

## Notes

- The proxy is GET-only and intended for read operations.
- Upstream responses are streamed back with the upstream status code.
- Graceful shutdown is implemented for SIGINT/SIGTERM.
