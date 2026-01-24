# HTTP from TCP

A minimal HTTP/1.1 server implementation built from scratch using raw TCP sockets in Go. This project demonstrates how to parse HTTP requests and write HTTP responses without using Go's standard `net/http` package.

## Overview

This project implements a complete HTTP/1.1 server by:
- Parsing HTTP requests directly from TCP streams
- Managing HTTP headers and request bodies
- Writing properly formatted HTTP responses
- Supporting advanced features like chunked transfer encoding and trailers

## Project Structure

```
.
├── cmd/
│   ├── httpserver/     # Full-featured HTTP server
│   └── tcplistener/    # Debug tool for inspecting HTTP requests
├── internal/
│   ├── headers/        # HTTP header parsing and management
│   ├── request/        # HTTP request parsing from TCP streams
│   ├── response/       # HTTP response writing
│   └── server/         # TCP server with connection handling
└── assets/
    └── vim.mp4         # Sample video file for testing
```

## Features

- ✅ **HTTP/1.1 Request Parsing**: Complete request line, headers, and body parsing
- ✅ **Streaming Parsing**: Handles partial/incomplete data from TCP streams
- ✅ **Chunked Transfer Encoding**: Supports chunked responses with trailers
- ✅ **Header Management**: Case-insensitive headers with support for repeated headers
- ✅ **Concurrent Connections**: One goroutine per connection
- ✅ **Graceful Shutdown**: Proper cleanup on SIGINT/SIGTERM
- ✅ **Error Handling**: Comprehensive error handling with appropriate HTTP status codes

## Building

```bash
# Build the HTTP server
go build -o httpserver ./cmd/httpserver

# Build the TCP listener (debug tool)
go build -o tcplistener ./cmd/tcplistener
```

## Running

### HTTP Server

```bash
./httpserver
```

The server listens on port `42069` by default.

### TCP Listener (Debug Tool)

```bash
./tcplistener
```

This tool listens on port `42069` and prints parsed HTTP request information to stdout.

## API Endpoints

The HTTP server (`httpserver`) implements the following routes:

- **`GET /`** - Returns a success HTML page (200 OK)
- **`GET /yourproblem`** - Returns 400 Bad Request
- **`GET /myproblem`** - Returns 500 Internal Server Error
- **`GET /httpbin/*`** - Proxies requests to httpbin.org with chunked encoding and trailers
  - Example: `GET /httpbin/get` proxies to `https://httpbin.org/get`
  - Returns chunked response with SHA256 hash in trailers
- **`GET /video`** - Serves the `vim.mp4` file with proper video content type

## Example Usage

```bash
# Test basic endpoint
curl http://localhost:42069/

# Test error endpoints
curl http://localhost:42069/yourproblem
curl http://localhost:42069/myproblem

# Test httpbin proxy
curl http://localhost:42069/httpbin/get

# Test video endpoint
curl http://localhost:42069/video -o output.mp4
```

## Architecture

### Request Parsing

The request parser uses a state machine to handle streaming data:
1. **RequestStateInit**: Parse the request line (method, target, version)
2. **HeadersState**: Parse HTTP headers
3. **BodyState**: Read body based on `Content-Length` header
4. **RequestStateDone**: Request fully parsed

### Response Writing

The response writer enforces proper HTTP response structure:
1. **Status Line**: HTTP version, status code, and status text
2. **Headers**: HTTP headers (with support for chunked encoding)
3. **Body**: Response body (or chunked data)
4. **Trailers**: Optional trailer headers (for chunked encoding)

### Headers

Headers are stored case-insensitively and support:
- Multiple values for the same header name
- Validation of header field names per RFC 7230
- Streaming parsing for incomplete data

## Testing

Run the test suite:

```bash
go test ./...
```

The project includes comprehensive tests for:
- Header parsing with various edge cases
- Request parsing with different chunk sizes
- Body parsing with and without Content-Length

## Implementation Details

- **No `net/http`**: All HTTP parsing and response writing is done manually
- **Raw TCP**: Uses `net.Listen` and `net.Conn` directly
- **Streaming**: Handles partial reads from TCP connections gracefully
- **Concurrent**: Each connection is handled in its own goroutine

## License

This is a learning project demonstrating low-level HTTP implementation.

