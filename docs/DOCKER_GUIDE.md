# Docker Setup Guide

## Quick Start

### 1. Build the Docker Image

```bash
docker build -t telos-matrix:latest .
```

### 2. Run with Your Telos File

```bash
# Place your telos.md in current directory
docker run -it \
  -v $(pwd)/telos.md:/config/telos.md:ro \
  -v telos-data:/data \
  telos-matrix:latest dump "Your idea"
```

### 3. Using Docker Compose (Recommended)

```bash
# Copy your telos.md to current directory
cp /path/to/your/telos.md .

# Start the container
docker-compose up -d

# Run commands
docker-compose exec telos-matrix dump "Your idea"
docker-compose exec telos-matrix review
docker-compose exec telos-matrix prune
```

## Volume Mapping

- `/config/telos.md` - Your Telos configuration (read-only)
- `/data` - Persistent database and idea storage
- `/logs` - Application logs

## Environment Variables

```bash
docker run \
  -e TELOS_FILE=/config/telos.md \
  -e RUST_LOG=debug \
  telos-matrix:latest dump "idea"
```

## Advanced: Custom Ollama Integration

If you want to use local Ollama:

```yaml
# docker-compose.yml with Ollama
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-models:/root/.ollama

  telos-matrix:
    build: .
    depends_on:
      - ollama
    environment:
      - OLLAMA_HOST=http://ollama:11434
    volumes:
      - ./telos.md:/config/telos.md:ro
      - telos-data:/data

volumes:
  ollama-models:
  telos-data:
```

## Troubleshooting

**Permission denied errors:**
```bash
docker-compose exec -u root telos-matrix chmod 777 /data
```

**Database locked:**
Ensure only one container instance is running:
```bash
docker-compose down
docker-compose up -d
```