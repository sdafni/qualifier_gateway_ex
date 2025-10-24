# Gateway Server

A simple Go-based gateway server that routes HTTP requests based on a custom header.

## Features

- Receives requests at `/gateway_endpoint`
- Routes to `url1/endpoint` or `url2/endpoint` based on the `X-Route-To` header
- Forwards all HTTP methods (GET, POST, PUT, DELETE, etc.)
- Preserves request headers and body
- Returns the upstream response to the client

## Configuration

The gateway can be configured using environment variables:

- `GATEWAY_URL1`: Target URL for route "url1" (default: `http://localhost:8081`)
- `GATEWAY_URL2`: Target URL for route "url2" (default: `http://localhost:8084`)
- `GATEWAY_PORT`: Port to run the gateway server on (default: `8080`)

## Usage

### Running the server

```bash
# With default configuration
go run main.go

# With custom URLs
GATEWAY_URL1=http://service1.example.com GATEWAY_URL2=http://service2.example.com go run main.go

# With custom port
GATEWAY_PORT=3000 go run main.go
```

### Building the binary

```bash
go build -o gateway
./gateway
```

### Making requests

The gateway expects a custom header `X-Route-To` with value `url1` or `url2`:

```bash
# Route to url1/endpoint
curl -H "X-Route-To: url1" http://localhost:8080/gateway_endpoint

# Route to url2/endpoint
curl -H "X-Route-To: url2" http://localhost:8080/gateway_endpoint

# POST request with body
curl -X POST -H "X-Route-To: url1" -H "Content-Type: application/json" \
  -d '{"key":"value"}' http://localhost:8080/gateway_endpoint
```

## Testing locally

To test the gateway locally, you can start simple backend servers:

### Terminal 1 - Start backend service 1
```bash
# Simple HTTP server on port 8081
python3 -m http.server 8081
```

### Terminal 2 - Start backend service 2
```bash
# Simple HTTP server on port 8084
python3 -m http.server 8084
```

### Terminal 3 - Start gateway
```bash
go run main.go
```

### Terminal 4 - Test requests
```bash
# Test routing to url1
curl -H "X-Route-To: url1" http://localhost:8080/gateway_endpoint

# Test routing to url2
curl -H "X-Route-To: url2" http://localhost:8080/gateway_endpoint
```

## Error Handling

- Returns `404 Not Found` for paths other than `/gateway_endpoint`
- Returns `400 Bad Request` if `X-Route-To` header is missing or invalid
- Returns `502 Bad Gateway` if the upstream server cannot be reached
- Returns `500 Internal Server Error` for configuration errors

## Request Flow

1. Client sends request to `http://localhost:8080/gateway_endpoint` with `X-Route-To` header
2. Gateway reads the `X-Route-To` header value
3. Gateway forwards the request to the appropriate backend:
   - `url1` → `{GATEWAY_URL1}/endpoint`
   - `url2` → `{GATEWAY_URL2}/endpoint`
4. Gateway receives the response from the backend
5. Gateway forwards the response back to the client
