# Deployment Guide

## Docker Compose (Production)

### 1. Configure Environment

Create a `.env` file:

```bash
DB_PASSWORD=your-secure-password
REDIS_PASSWORD=your-secure-password
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
JWT_SECRET=your-jwt-secret-min-32-chars
DOMAIN=chat.example.com
```

### 2. Deploy

```bash
make prod
```

This starts all services with:
- Caddy reverse proxy with auto TLS (Let's Encrypt)
- PostgreSQL with persistent storage
- Redis with authentication
- MinIO for file storage
- API server with health checks

### 3. Verify

```bash
curl https://chat.example.com/api/v1/health
```

## Security Checklist

- [ ] Change all default passwords
- [ ] Set a strong JWT_SECRET (min 32 characters)
- [ ] Configure your domain in DOMAIN env var
- [ ] Review rate limiting settings
- [ ] Enable firewall (only expose ports 80/443)
- [ ] Set up automated backups for PostgreSQL

## Backup

```bash
# Database backup
docker compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U feather feather > backup.sql

# Restore
cat backup.sql | docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U feather feather
```
