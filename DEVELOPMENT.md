# Development Environment Setup

This guide will help you set up a development environment for Semyi.

## Prerequisites

- Go 1.24 or later
- Node.js 20.19 or later
- pnpm
- Docker and Docker Compose
- Make

## Getting Started

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/teknologi-umum/semyi.git
   cd semyi
   ```

2. **Install Frontend Dependencies**:
   ```bash
   cd frontend
   pnpm install
   ```

3. **Install Backend Dependencies**:
   ```bash
   cd backend
   go mod download
   ```

## Development Setup

### Frontend Development

1. **Start the Development Server**:
   ```bash
   cd frontend
   cp .env.example .env
   pnpm dev
   ```
   This will start the development server with hot reloading enabled.

### Backend Development

1. **Build the Backend**:
   ```bash
   cd backend
   go build
   ```

2. **Run the Backend**:
   ```bash
   ./semyi
   ```
   The backend will use the default configuration file at `config.json`.

You can run an optional uptime server that mimics a bad application:
```bash
go run uptime-server/golang/main.go
```

## Configuration

Create a `config.json` file in the root directory with the following structure:

```json
{
  "alerting": {
    "telegram": {
      "enabled": true,
      "url": "https://api.telegram.org",
      "chat_id": "123456789"
    }
  },
  "monitors": [
    {
      "unique_id": "monitor-1",
      "name": "Example Monitor",
      "type": "http",
      "interval": 30,
      "timeout": 30,
      "http_endpoint": "https://example.com/_healthz"
    }
  ]
}
```

## Environment Variables

Set the following environment variables for development:

```bash
export ENVIRONMENT=development
export CONFIG_PATH="../config.json"
export DB_PATH="../db.duckdb"
export STATIC_PATH=frontend/dist
export PORT=5000
```

## Testing

### Frontend Tests

```bash
cd frontend
pnpm test
```

### Backend Tests

```bash
cd backend
go test ./...
```

## Building for Production

To build the application for production:

```bash
make build
```

This will create a production-ready Docker image.

## Troubleshooting

If you encounter any issues:

1. Check that all dependencies are installed correctly
2. Ensure the configuration file is properly formatted
3. Check the logs for any error messages
4. Make sure all required environment variables are set

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Node.js Documentation](https://nodejs.org/en/docs/)
- [Docker Documentation](https://docs.docker.com/) 