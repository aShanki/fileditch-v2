# File Hosting Service

A secure, containerized file hosting service with automatic file expiration and password protection.

## Features

- File upload with size limits
- Configurable file storage duration (1 hour, 1 day, 7 days, 30 days, permanent)
- Password protection for individual files
- Automatic file cleanup based on expiration
- Mobile-responsive interface
- Simple file management for uploaders

## Prerequisites

- Docker
- Docker Compose

## Quick Start

1. Clone this repository
2. Configure environment variables (optional)
3. Run the application:
   ```bash
   docker-compose up --build
   ```
4. Access the application at http://localhost

## Configuration

The following environment variables can be configured in the docker-compose.yml:

- `MONGODB_URI`: MongoDB connection string
- `JWT_SECRET`: Secret key for JWT tokens
- `MAX_FILE_SIZE`: Maximum allowed file size in bytes (default: 100MB)

## Cloud Storage Integration

The application is designed to work with cloud storage providers. To integrate with your preferred provider:

1. Add your cloud storage credentials to the backend environment variables
2. Implement the cloud storage interface in `backend/storage/cloud.go`

## Security Considerations

- Always use HTTPS in production
- Change the default JWT secret
- Configure appropriate file size limits
- Use strong passwords for file protection
- Regularly update dependencies

## Development

To modify the application:

1. Frontend code is in the `frontend/` directory
2. Backend Go code is in the `backend/` directory
3. Rebuild containers after changes:
   ```bash
   docker-compose up --build
   ```

## License

MIT