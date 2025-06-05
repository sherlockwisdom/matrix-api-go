# Matrix API Go

A Go-based Matrix server API implementation that provides user management, messaging, and websocket functionality.

## Features

- User Management
  - User registration
  - User login
  - Access token management
- Messaging
  - Send messages to Matrix rooms
  - Support for different platforms
- Device/Bridge Management
  - Add devices for different platforms
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

## API Endpoints

### User Management

- `POST /` - Create a new user
- `POST /login` - Login existing user

### Messaging

- `POST /{platform}/message/{roomid}` - Send a message to a specific room

### Device Management

- `POST /{platform}/devices/` - Add a new device/bridge for a platform

## Running the Application

### Command Line Arguments

The application supports several command-line arguments:

- `--create` - Create a new user
- `--login <username>` - Login as a specific user
- `--websocket` - Start the websocket server

### Starting the Server

```bash
go run main.go
```

The server will start on the configured host and port. If TLS certificates are provided, it will run in HTTPS mode.

## API Documentation

Swagger documentation is available at `/swagger/*` when the server is running.

## Security

- The application supports TLS encryption
- Access tokens are required for authenticated operations
- Passwords are handled securely

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details. 