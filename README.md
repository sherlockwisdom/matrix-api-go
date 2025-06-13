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
- Swagger Documentation
  - API documentation available at `/swagger/*`

## Prerequisites

- Go 1.x
- Matrix server instance
- TLS certificates (for HTTPS support)

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

## API Endpoints

### User Management

- `POST /` - Create a new user
- `POST /login` - Login existing user

### Messaging

- `POST /{platform}/message/{contact}` - Send a message to a contact
  - `platform`: Supported platform (e.g., 'wa' for WhatsApp)
  - `contact`: E.164 phone number (without '+' prefix)

### Bridge Management

- `POST /{platform}/devices` - Add a new bridge for a platform
  - Returns WebSocket URL for real-time communication

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

## API Documentation

Swagger documentation is available at `/swagger/*` when the server is running.

## Security

- The application supports TLS encryption
- Access tokens are required for authenticated operations
- Passwords are handled securely
- Input validation for usernames, passwords, and phone numbers
- CORS support with configurable origins

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details. 