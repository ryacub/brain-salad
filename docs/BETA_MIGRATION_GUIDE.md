# Beta Migration Guide - Telos Idea Matrix

Welcome to the Telos Idea Matrix beta! This guide will help you migrate from any previous version or get started fresh with the beta release.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Migration Paths](#migration-paths)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Data Migration](#data-migration)
6. [Verification](#verification)
7. [Troubleshooting](#troubleshooting)
8. [Rollback](#rollback)

---

## Prerequisites

Before starting the migration, ensure you have:

- **Docker** (v20.10+) and **Docker Compose** (v2.0+) installed
- **Node.js** (v20+) if running the frontend separately
- **Go** (v1.24+) if building from source
- At least **1GB** of free disk space
- **Backup** of your existing data (if migrating from a previous version)

### System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 1 core | 2+ cores |
| RAM | 512MB | 1GB+ |
| Disk | 500MB | 2GB+ |
| OS | Linux, macOS, Windows (WSL2) | Linux |

---

## Migration Paths

### Path 1: Fresh Installation (New Users)

If you're new to Telos Idea Matrix:

1. Skip to [Installation](#installation)
2. Follow [Configuration](#configuration)
3. Start using the tool!

### Path 2: Migrating from Rust CLI Version

If you're using the older Rust-based CLI:

1. **Backup your data**:
   ```bash
   # Backup your database
   cp ~/.telos/ideas.db ~/.telos/ideas.db.backup

   # Backup your configuration
   cp telos.md telos.md.backup
   ```

2. **Export your ideas** (optional, for extra safety):
   ```bash
   # Using the old CLI
   tm export --format json > ideas_backup.json
   ```

3. **Stop the old version**:
   ```bash
   # If running as a service
   docker-compose down

   # Or kill any running processes
   pkill tm
   ```

4. **Install the new version** - see [Installation](#installation)

5. **Migrate your data** - see [Data Migration](#data-migration)

### Path 3: Upgrading from Previous Beta

If you're already on a beta version:

1. **Backup your data**:
   ```bash
   ./scripts/deploy.sh backup
   ```

2. **Pull the latest changes**:
   ```bash
   git pull origin main
   ```

3. **Rebuild and restart**:
   ```bash
   ./scripts/deploy.sh build --env staging
   ./scripts/deploy.sh restart --env staging
   ```

---

## Installation

### Option 1: Docker Compose (Recommended)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/rayyacub/telos-idea-matrix.git
   cd telos-idea-matrix
   ```

2. **Set up environment**:
   ```bash
   cp .env.staging .env
   ```

3. **Deploy to staging**:
   ```bash
   ./scripts/deploy.sh deploy --env staging
   ```

4. **Verify installation**:
   ```bash
   ./scripts/smoke-test.sh
   ```

### Option 2: Manual Installation

1. **Build the Go binaries**:
   ```bash
   cd go
   make build
   ```

2. **Build the frontend**:
   ```bash
   cd web
   npm install
   npm run build
   ```

3. **Run the services**:
   ```bash
   # Terminal 1: API server
   ./go/cli web

   # Terminal 2: Frontend (development mode)
   cd web
   npm run dev
   ```

---

## Configuration

### 1. Create Your Telos File

The `telos.md` file defines your personal goals and evaluation criteria:

```bash
cat > telos.md << 'EOF'
# My Telos

## Goals
- G1: Launch a profitable SaaS product (Deadline: 2025-12-31)
- G2: Build a personal brand through open source (Deadline: 2025-06-30)
- G3: Improve technical skills in Go and systems programming (Deadline: 2025-09-30)

## Strategies
- S1: Ship early and often, iterate based on feedback
- S2: Focus on one technology stack to maximize depth
- S3: Build in public to maintain accountability
- S4: Prioritize projects with clear revenue potential

## Stack
- Primary: Go, TypeScript, PostgreSQL, SvelteKit
- Secondary: Docker, Kubernetes, GitHub Actions
- Avoiding: Java, PHP (unless absolutely necessary)

## Failure Patterns
- Context switching: Starting new projects before finishing current ones
- Perfectionism: Over-engineering solutions before validating market fit
- Tutorial hell: Watching tutorials instead of building
- Shiny object syndrome: Chasing new technologies without mastering current ones
EOF
```

### 2. Environment Configuration

Edit `.env` to match your setup:

```bash
# API Configuration
API_URL=http://localhost:8080
PORT=8080

# Frontend
VITE_API_URL=http://localhost:8080

# Logging
LOG_LEVEL=info
```

---

## Data Migration

### Automatic Migration

The new version includes automatic database migration:

1. **Place your old database** in the expected location:
   ```bash
   mkdir -p ~/.telos
   cp ~/.telos/ideas.db.backup ~/.telos/ideas.db
   ```

2. **Start the application** - migrations run automatically:
   ```bash
   ./scripts/deploy.sh start --env staging
   ```

3. **Verify migration**:
   ```bash
   # Check logs
   docker logs telos-api-staging

   # Test API
   curl http://localhost:8080/ideas
   ```

### Manual Migration

If automatic migration fails:

1. **Export from old version**:
   ```bash
   # Using the old CLI
   tm export --format json > ideas_export.json
   ```

2. **Import to new version**:
   ```bash
   # Using the new CLI
   ./go/cli import ideas_export.json
   ```

### Data Compatibility

| Old Version | New Version | Migration Path |
|-------------|-------------|----------------|
| Rust CLI v0.1.x | Go v1.0 Beta | Automatic |
| Rust CLI v0.2.x | Go v1.0 Beta | Automatic |
| Custom/External | Go v1.0 Beta | Manual (JSON import) |

---

## Verification

### 1. Run Smoke Tests

```bash
./scripts/smoke-test.sh --api-url http://localhost:8080
```

Expected output:
```
✓ API health check passed
✓ API version endpoint working
✓ Frontend is accessible
✓ Ideas list endpoint working
✓ Create idea endpoint working
✓ Telos config endpoint working
✓ Database persistence working
```

### 2. Manual Verification

#### Test CLI:
```bash
# Dump a test idea
docker exec telos-api-staging tm dump "Test idea for beta"

# Review ideas
docker exec telos-api-staging tm review --limit 5
```

#### Test Web Interface:
1. Open browser to `http://localhost:3000`
2. Verify you can see your ideas
3. Try creating a new idea
4. Check that scoring works correctly

#### Test API:
```bash
# Get all ideas
curl http://localhost:8080/ideas

# Get telos config
curl http://localhost:8080/telos

# Health check
curl http://localhost:8080/health
```

### 3. Check Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/staging_admin_2024)

---

## Troubleshooting

### Issue: Database Migration Fails

**Solution**:
```bash
# Check logs
docker logs telos-api-staging

# Manually run migration
docker exec telos-api-staging tm migrate --verbose

# If all else fails, start fresh
rm -rf ~/.telos/ideas.db
docker restart telos-api-staging
```

### Issue: Frontend Can't Connect to API

**Solution**:
```bash
# Check API is running
curl http://localhost:8080/health

# Verify environment variables
docker exec telos-frontend-staging env | grep VITE_API_URL

# Rebuild frontend with correct API URL
cd web
VITE_API_URL=http://localhost:8080 npm run build
```

### Issue: Port Already in Use

**Solution**:
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process or change port in .env
PORT=8081 ./scripts/deploy.sh restart --env staging
```

### Issue: Docker Compose Errors

**Solution**:
```bash
# Clean up and rebuild
docker-compose -f docker-compose.staging.yml down -v
docker-compose -f docker-compose.staging.yml build --no-cache
docker-compose -f docker-compose.staging.yml up -d
```

---

## Rollback

If you need to rollback to the previous version:

### Rollback Data

```bash
# Restore database backup
cp ~/.telos/ideas.db.backup ~/.telos/ideas.db

# Restore telos config
cp telos.md.backup telos.md
```

### Rollback Application

```bash
# Stop new version
./scripts/deploy.sh stop --env staging

# Restore old version (if you have it)
# ... (specific to your previous setup)
```

### Using Backup Script

```bash
# List available backups
ls -la backups/

# Restore specific backup
./scripts/deploy.sh restore backups/20241119_120000/
```

---

## Getting Help

### Resources

- **Documentation**: https://github.com/rayyacub/telos-idea-matrix/tree/main/docs
- **Issues**: https://github.com/rayyacub/telos-idea-matrix/issues
- **Discussions**: https://github.com/rayyacub/telos-idea-matrix/discussions

### Reporting Issues

When reporting issues, please include:

1. **Version information**:
   ```bash
   docker exec telos-api-staging tm --version
   ```

2. **Error logs**:
   ```bash
   docker logs telos-api-staging --tail 100
   ```

3. **System information**:
   ```bash
   docker --version
   docker-compose --version
   uname -a
   ```

4. **Steps to reproduce** the issue

---

## Next Steps

After successful migration:

1. ✅ **Explore the new web interface** at http://localhost:3000
2. ✅ **Set up monitoring** - check Grafana dashboards
3. ✅ **Provide feedback** - see [BETA_FEEDBACK.md](./BETA_FEEDBACK.md)
4. ✅ **Join the community** - participate in discussions
5. ✅ **Report bugs** - help us improve for v1.0!

---

**Thank you for being a beta tester!** Your feedback is invaluable in making Telos Idea Matrix better for everyone.
