# Matrix API Go

A Go-based Matrix messaging bridge API that enables seamless communication across different messaging platforms. It provides endpoints for user management, message sending, and platform bridging capabilities.

## Features

- User Management
  - User registration
  - User login
  - Access token management
- Messaging
  - Send messages to contacts using E.164 phone number format
  - Support for multiple messaging platforms
- Platform Bridge Management
  - Add bridges for different platforms (WhatsApp, Signal)
  - WebSocket support for real-time communication
- Interactive API Documentation
  - Swagger UI available at `/docs` when server is running

## Prerequisites

- Go 1.x
- Matrix server instance
- TLS certificates (for HTTPS support)
- Python 3.x (for documentation generation)

## Installation

### Go Dependencies

```bash
go mod download
```

### Documentation Dependencies

The project includes comprehensive documentation built with Sphinx. To set up the documentation environment:

1. Navigate to the tutorials directory:
```bash
cd tutorials
```

2. Install Python dependencies:
```bash
pip install -r requirements.txt
```

3. Build the documentation:
```bash
make html
```

The built documentation will be available in `tutorials/_build/html/`.

## Configuration

1. Copy the example configuration file:
```bash
cp conf.yaml.example conf.yaml
```

2. Configure the following in `conf.yaml`:
   - Server settings (host, port, TLS)
   - Matrix homeserver details
   - Bridge configurations for supported platforms
   - Keystore filepath
   - Default user credentials

## API Documentation

When the server is running, you can access the interactive API documentation at:
- **Swagger UI**: `http://localhost:8080/docs` (or your configured host/port)

The documentation provides:
- Complete API endpoint reference
- Interactive testing interface
- Request/response examples
- Authentication details

## WebSocket Support

The API provides WebSocket endpoints for real-time communication:
- WebSocket URL format: `/ws/{platform}/{username}`
- Supports secure WebSocket connections (WSS) when TLS is enabled
- Handles real-time message synchronization

## Running the Application

### Starting the Server

```bash
go run main.go
```

The server will start on the configured host and port. If TLS certificates are provided, it will run in HTTPS mode.

### WebSocket Server

The WebSocket server runs on port 8090 by default:
- HTTP: `ws://localhost:8090`
- HTTPS: `wss://localhost:8090`

### Documentation Server

To serve the built documentation locally:

```bash
cd tutorials/_build/html
python -m http.server 8000
```

Then visit `http://localhost:8000` to view the documentation.

## Development

### Building Documentation

To rebuild the documentation after making changes:

```bash
cd tutorials
make clean
make html
```

### Documentation Structure

- `tutorials/` - Sphinx documentation source
- `tutorials/_build/` - Generated documentation output
- `tutorials/requirements.txt` - Python dependencies for documentation
- `tutorials/Makefile` - Build commands for documentation

## Security

- The application supports TLS encryption
- Access tokens are required for authenticated operations
- Passwords are handled securely
- Input validation for usernames, passwords, and phone numbers
- CORS support with configurable origins

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details. 