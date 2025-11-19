#!/bin/bash
#
# Deployment script for Telos Idea Matrix
# Supports Docker, local, and production deployments
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_usage() {
    cat <<EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build       Build Docker images and binaries
    deploy      Deploy the application
    start       Start the services
    stop        Stop the services
    restart     Restart the services
    logs        Show service logs
    test        Run all tests
    clean       Clean build artifacts
    help        Show this help message

Options:
    --env ENV       Environment: dev, staging, prod (default: dev)
    --port PORT     API server port (default: 8080)
    --no-cache      Build without Docker cache
    --verbose       Enable verbose output

Examples:
    $0 build
    $0 deploy --env prod
    $0 start --port 3000
    $0 logs --env prod
EOF
}

# Parse arguments
COMMAND=${1:-help}
shift || true

ENV="dev"
PORT="8080"
NO_CACHE=""
VERBOSE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --env)
            ENV="$2"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        --verbose)
            VERBOSE="-v"
            set -x
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate environment
if [[ ! "$ENV" =~ ^(dev|staging|prod)$ ]]; then
    log_error "Invalid environment: $ENV"
    exit 1
fi

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Environment-specific configuration
case $ENV in
    dev)
        COMPOSE_FILE="docker-compose.yml"
        ;;
    staging)
        COMPOSE_FILE="docker-compose.staging.yml"
        ;;
    prod)
        COMPOSE_FILE="docker-compose.prod.yml"
        ;;
esac

# Commands

cmd_build() {
    log_info "Building Telos Idea Matrix..."

    # Build Go binaries
    log_info "Building Go binaries..."
    cd go
    make build
    cd ..

    # Build frontend
    log_info "Building frontend..."
    cd web
    npm install
    npm run build
    cd ..

    # Build Docker images
    log_info "Building Docker images..."
    docker-compose -f "$COMPOSE_FILE" build $NO_CACHE

    log_info "Build complete!"
}

cmd_deploy() {
    log_info "Deploying to $ENV environment..."

    # Pre-deployment checks
    log_info "Running pre-deployment checks..."

    # Check if telos.md exists
    if [[ ! -f "telos.md" ]]; then
        log_error "telos.md not found. Create one before deploying."
        exit 1
    fi

    # Run tests
    log_info "Running tests..."
    cmd_test

    # Build
    cmd_build

    # Deploy based on environment
    case $ENV in
        dev)
            log_info "Starting development environment..."
            docker-compose -f "$COMPOSE_FILE" up -d
            ;;
        staging|prod)
            log_info "Deploying to $ENV..."
            # In production, you might use a different orchestration system
            docker-compose -f "$COMPOSE_FILE" up -d
            ;;
    esac

    log_info "Deployment complete!"
    log_info "API available at: http://localhost:$PORT"
}

cmd_start() {
    log_info "Starting services..."

    if [[ ! -f "$COMPOSE_FILE" ]]; then
        log_error "Compose file not found: $COMPOSE_FILE"
        exit 1
    fi

    docker-compose -f "$COMPOSE_FILE" up -d

    log_info "Services started!"
    docker-compose -f "$COMPOSE_FILE" ps
}

cmd_stop() {
    log_info "Stopping services..."
    docker-compose -f "$COMPOSE_FILE" down
    log_info "Services stopped!"
}

cmd_restart() {
    log_info "Restarting services..."
    cmd_stop
    cmd_start
}

cmd_logs() {
    log_info "Showing logs for $ENV environment..."
    docker-compose -f "$COMPOSE_FILE" logs -f
}

cmd_test() {
    log_info "Running tests..."

    # Go tests
    log_info "Running Go tests..."
    cd go
    make test

    # Integration tests
    log_info "Running integration tests..."
    make test-integration

    # Go back to project root
    cd ..

    # Frontend tests
    log_info "Running frontend tests..."
    cd web
    npm test -- --run
    cd ..

    # Security scans
    log_info "Running security scans..."
    cd go
    if command -v gosec &> /dev/null; then
        ~/go/bin/gosec ./... || log_warn "gosec found some issues"
    else
        log_warn "gosec not installed, skipping security scan"
    fi
    cd ..

    cd web
    npm audit --production
    cd ..

    log_info "All tests passed!"
}

cmd_clean() {
    log_info "Cleaning build artifacts..."

    # Clean Go binaries
    cd go
    make clean
    cd ..

    # Clean frontend build
    rm -rf web/build
    rm -rf web/node_modules/.vite

    # Clean Docker images (optional)
    read -p "Remove Docker images? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f "$COMPOSE_FILE" down --rmi all
    fi

    log_info "Clean complete!"
}

cmd_status() {
    log_info "Service status:"
    docker-compose -f "$COMPOSE_FILE" ps

    log_info ""
    log_info "Resource usage:"
    docker stats --no-stream
}

cmd_backup() {
    log_info "Creating backup..."

    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"

    # Backup database
    if [[ -f "$HOME/.telos/ideas.db" ]]; then
        cp "$HOME/.telos/ideas.db" "$BACKUP_DIR/ideas.db"
        log_info "Database backed up to $BACKUP_DIR/ideas.db"
    fi

    # Backup telos configuration
    if [[ -f "telos.md" ]]; then
        cp telos.md "$BACKUP_DIR/telos.md"
        log_info "Telos config backed up to $BACKUP_DIR/telos.md"
    fi

    log_info "Backup complete!"
}

cmd_restore() {
    if [[ -z "$1" ]]; then
        log_error "Please specify backup directory"
        echo "Usage: $0 restore <backup_dir>"
        exit 1
    fi

    BACKUP_DIR="$1"

    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_error "Backup directory not found: $BACKUP_DIR"
        exit 1
    fi

    log_warn "This will overwrite current data. Continue? (y/N)"
    read -p "" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Restore cancelled"
        exit 0
    fi

    # Restore database
    if [[ -f "$BACKUP_DIR/ideas.db" ]]; then
        mkdir -p "$HOME/.telos"
        cp "$BACKUP_DIR/ideas.db" "$HOME/.telos/ideas.db"
        log_info "Database restored"
    fi

    # Restore telos configuration
    if [[ -f "$BACKUP_DIR/telos.md" ]]; then
        cp "$BACKUP_DIR/telos.md" telos.md
        log_info "Telos config restored"
    fi

    log_info "Restore complete!"
}

# Main command dispatcher
case $COMMAND in
    build)
        cmd_build
        ;;
    deploy)
        cmd_deploy
        ;;
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    logs)
        cmd_logs
        ;;
    test)
        cmd_test
        ;;
    clean)
        cmd_clean
        ;;
    status)
        cmd_status
        ;;
    backup)
        cmd_backup
        ;;
    restore)
        cmd_restore "$@"
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        log_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac
