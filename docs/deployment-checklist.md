# Kabletown Deployment Checklist

This checklist guides you through deploying Kabletown microservices to production.

---

## Pre-Deployment Preparation

### Infrastructure Requirements

- [ ] **Hardware**:
  - [ ] CPU: 4+ cores (recommended: 8+ for transcoding)
  - [ ] RAM: 8GB minimum, 16GB+ recommended
  - [ ] Storage: SSD for database, HDD for media files
  - [ ] Network: Gigabit Ethernet (1 Gbps minimum)

- [ ] **Operating System**:
  - [ ] Ubuntu 22.04 LTS (tested)
  - [ ] Alternative: Debian 11+, CentOS 8+

- [ ] **Docker**:
  - [ ] Docker Engine 24.0+
  - [ ] Docker Compose v2.0+
  - [ ] Verify: `docker --version`, `docker compose version`

- [ ] **Database**:
  - [ ] MySQL 8.0+ (or MariaDB 10.6+)
  - [ ] Disk space for database (estimate: 1GB per 10k items)
  - [ ] Automated backup strategy in place

### DNS & Networking

- [ ] **Domain Configuration**:
  - [ ] DNS A record pointing to server IP
  - [ ] HTTPS certificate obtained (Let's Encrypt recommended)
  - [ ] Subdomain for each service (optional: one subdomain per service)

- [ ] **Firewall Rules**:
  - [ ] Port 80 open (HTTP for certbot)
  - [ ] Port 443 open (HTTPS)
  - [ ] Port 3306: ONLY internal access (database)
  - [ ] Port 8081, 8004, 8005: Blocked from external (internal only)

- [ ] **SSL/TLS**:
  - [ ] Certbot installed
  - [ ] Certificate auto-renewal configured
  - [ ] TLS 1.2+ enforced

---

## Database Setup

### Initial Database Creation

```bash
# 1. Connect as root
mysql -u root -p

# 2. Create database
CREATE DATABASE jellyfin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 3. Create user
CREATE USER 'jellyfin'@'localhost' IDENTIFIED BY 'CHANGE_THIS_PASSWORD';
GRANT ALL PRIVILEGES ON jellyfin.* TO 'jellyfin'@'localhost';
FLUSH PRIVILEGES;

# 4. Exit
EXIT;
```

- [ ] Database created: `jellyfin`
- [ ] User created: `jellyfin`
- [ ] Password changed from default
- [ ] User restricted to localhost or specific IP

### Import Schema

```bash
# Import initial schema
mysql -u jellyfin -p jellyfin < Kabletown/migrations/schema.sql

# Verify tables created
mysql -u jellyfin -p -e "USE jellyfin; SHOW TABLES;"
```

- [ ] Schema imported successfully
- [ ] Tables verified: 7+ tables created
- [ ] Default admin user created (`admin` / `changeme`)

### Database Optimization

```sql
-- MySQL performance tuning (add to /etc/mysql/mysql.conf.d/kabletown.cnf)
[mysqld]
# Connection pooling
max_connections = 100
innodb_buffer_pool_size = 2G  # 25% of system RAM

# Query cache (MySQL 5.7) or read buffer (8.0)
sort_buffer_size = 4M
read_buffer_size = 2M
read_rnd_buffer_size = 1M

# InnoDB settings
innodb_log_file_size = 512M
innodb_flush_log_at_trx_commit = 2  # Better performance, minor durability trade-off
innodb_file_per_table = 1

# Query logging (disable in production if not debugging)
slow_query_log = ON
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

- [ ] Performance tuning applied
- [ ] MySQL restarted
- [ ] Settings verified: `SHOW VARIABLES LIKE 'innodb_buffer_pool_size';`

---

## Docker Compose Setup

### Environment Configuration

Create `.env` file in project root:

```bash
# Database
DB_HOST=mysql
DB_PORT=3306
DB_USER=jellyfin
DB_PASSWORD=CHANGE_THIS_PASSWORD
DB_NAME=jellyfin
DB_ROOT_PASSWORD=CHANGE_THIS_ROOT_PASSWORD

# Service ports
AUTH_SERVICE_PORT=8081
ITEM_SERVICE_PORT=8004
STREAMING_SERVICE_PORT=8005
NGINX_HTTP_PORT=80
NGINX_HTTPS_PORT=443

# Paths
MEDIA_DIR=/media
TRANSCODE_DIR=/transcode
DATA_DIR=/var/lib/jellyfin

# SSL
SSL_CERT_PATH=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

- [ ] `.env` file created
- [ ] All passwords changed
- [ ] SSL cert paths verified
- [ ] File permissions: `chmod 600 .env`

### Directory Structure

```bash
# Create required directories
mkdir -p /var/lib/jellyfin
mkdir -p /media/movies
mkdir -p /media/tv
mkdir -p /transcode
mkdir -p /backups/mysql

# Set permissions
chown -R 1000:1000 /var/lib/jellyfin  # UID 1000 from Dockerfile
chown -R 1000:1000 /transcode
chmod 755 /media /transcode
```

- [ ] Directories created
- [ ] Permissions set correctly
- [ ] Disk space verified:
  - [ ] `/media`: Media library size + 20% buffer
  - [ ] `/transcode`: At least 50GB (ephemeral)
  - [ ] `/var/lib/jellyfin`: Metadata + cache (10GB+)

### Docker Compose Configuration

```yaml
# docker-compose.yml (production-ready snippet)
services:
  mysql:
    image: mysql:8.0
    container_name: kabletown-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - kabletown
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 60s

  auth-service:
    build:
      context: ./auth-service
      dockerfile: Dockerfile
    container_name: kabletown-auth-service
    restart: unless-stopped
    environment:
      - PORT=8081
      - DB_HOST=mysql
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
    networks:
      - kabletown
    depends_on:
      mysql:
        condition: service_healthy

  # ... item-service, streaming-service, nginx ...
```

- [ ] Docker Compose file created
- [ ] Health checks configured for MySQL
- [ ] Restart policies set (`unless-stopped`)
- [ ] Environment variables injected from `.env`

---

## Nginx Reverse Proxy Configuration

### SSL Setup (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtain certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Verify auto-renewal
sudo certbot renew --dry-run
```

- [ ] SSL certificate obtained
- [ ] Auto-renewal working
- [ ] Certificate path verified

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/kabletown

server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Timeouts
    client_body_timeout 300s;
    send_timeout 300s;
    proxy_read_timeout 300s;

    # Increase buffer sizes for video streaming
    client_max_body_size 100M;
    client_body_buffer_size 128M;

    # Auth Service
    location /auth {
        proxy_pass http://auth-service:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
    }

    # Item Service
    location /items {
        proxy_pass http://item-service:8004;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
    }

    # Streaming Service (HLS + Progressive)
    location /videos {
        proxy_pass http://streaming-service:8005;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        
        # Required for streaming
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding on;
        tcp_nopush on;
        tcp_nodelay on;
    }
}
```

- [ ] Nginx config created
- [ ] SSL paths verified
- [ ] Reverse proxy rules tested
- [ ] Security headers configured
- [ ] Streaming-specific settings (proxy_buffering off) set

### Test Nginx

```bash
# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Verify services
curl https://yourdomain.com/auth/health
# Should return: {"status":"ok"}
```

- [ ] Nginx configuration tested
- [ ] Services accessible via HTTPS
- [ ] SSL working (no certificate warnings)

---

## Service-Specific Configuration

### Auth Service

- [ ] Database connection verified
- [ ] Default admin user exists
- [ ] Password changed from `changeme`
- [ ] Device registration functional
- [ ] Token validation functional

**Test:**
```bash
curl -X POST https://yourdomain.com/auth/Sessions \
  -H "Content-Type: application/json" \
  -d '{"Username":"admin","Pw":"newpassword"}'

# Should return AccessToken
```

### Item Service

- [ ] Database connection verified
- [ ] Tables created (base_items, item_values, user_data)
- [ ] Migration scripts applied
- [ ] Health endpoint responds

**Test:**
```bash
curl -H "X-Emby-Authorization: MediaBrowser Token=\"yourtoken\"" \
  https://yourdomain.com/items/health

# Should return {"status":"ok"}
```

### Streaming Service

- [ ] FFmpeg installed and accessible
- [ ] `/transcode` directory writable
- [ ] `/media` directory readable
- [ ] Health endpoint responds

**Verify FFmpeg:**
```bash
# Check FFmpeg path
which ffmpeg
ffmpeg -version

# Configure environment variable
export FFMPEG_PATH=/usr/bin/ffmpeg
```

- [ ] Test transcoding: Upload and transcode a short video file
- [ ] Verify segment creation in `/transcode`
- [ ] Verify cleanup after transcode ends

---

## Initial Data Migration (From Jellyfin)

### Backup Existing Jellyfin Database

```bash
# Export Jellyfin database
mysqldump -h jellyfin-host -u root -p jellyfin > jellyfin_backup_$(date +%Y%m%d).sql

# Verify backup
wc -l jellyfin_backup_*.sql
```

- [ ] Full database backup created
- [ ] Backup stored securely (offsite recommended)
- [ ] Backup file verified

### Import to Kabletown

```bash
# Import schema
mysql -h localhost -u jellyfin -p jellyfin < Kabletown/migrations/schema.sql

# Import data
mysql -h localhost -u jellyfin -p jellyfin < jellyfin_backup_20260313.sql

# Run migration scripts
mysql -h localhost -u jellyfin -p jellyfin < Kabletown/migrations/sync_from_jellyfin.sql

# Verify counts
mysql -h localhost -u jellyfin -p -e "
USE jellyfin;
SELECT 'Users' as Table, COUNT(*) as Count FROM users
UNION ALL
SELECT 'BaseItems', COUNT(*) FROM base_items
UNION ALL
SELECT 'UserData', COUNT(*) FROM user_data;
"
```

- [ ] Data imported successfully
- [ ] Row counts match original Jellyfin
- [ ] Item metadata visible in API
- [ ] User data (playstate) preserved

### Media Files Verification

```bash
# Mount or symlink media files
ln -s /mnt/jellyfin-media /media

# Verify media readable
ls -la /media/movies | head
ls -la /media/tv | head
```

- [ ] Media files accessible
- [ ] File permissions correct (readable by 1000:1000)
- [ ] Paths match database entries (or paths updated)

---

## Health Checks & Monitoring

### Health Check Endpoints

```bash
# Auth Service
curl https://yourdomain.com/auth/health

# Item Service
curl https://yourdomain.com/items/health

# Streaming Service
curl https://yourdomain.com/videos/health

# All via Nginx
curl https://yourdomain.com/health
```

- [ ] All services respond to health checks
- [ ] Status 200 for all
- [ ] JSON format valid

### Prometheus Metrics (Optional)

```
# Add metrics endpoint to each service
GET /metrics

# Sample metrics
transcoding_jobs_active{job="streaming"} 2
auth_tokens_created_total{job="auth"} 150
item_queries_total{job="item"} 5000
```

- [ ] Metrics endpoints available
- [ ] Prometheus configured to scrape
- [ ] Grafana dashboards created (optional)

### Log Aggregation

```bash
# Docker logs
docker logs -f kabletown-auth-service
docker logs -f kabletown-item-service
docker logs -f kabletown-streaming-service

# Syslog (if configured)
tail -f /var/log/kabletown/*.log
```

- [ ] Logging configured
- [ ] Log rotation set up
- [ ] Log aggregation tool chosen (ELK, Loki, etc.)

---

## Performance Tuning

### MySQL Tuning

```sql
-- Check slow queries
SELECT * FROM mysql.slow_log LIMIT 10;

-- Check index usage
EXPLAIN SELECT * FROM base_items WHERE TopParentId = 'abc123';

-- Add missing indexes if needed
CREATE INDEX idx_missing ON table(column);
```

- [ ] Slow query log analyzed
- [ ] Queries optimized with EXPLAIN
- [ ] Indexes added for slow queries

### FFmpeg Tuning

```bash
# Limit concurrent transcodes (in streaming service)
export MAX_CONCURRENT_TRANSCODES=2

# CPU priority
cgroupset -p /path/to/process -n realtime -p -m 2 -c 4

# I/O priority
ionice -c 1 -n 0 -p $(pgrep ffmpeg)
```

- [ ] Concurrent transcode limit set
- [ ] CPU priority configured (optional)
- [ ] I/O priority configured (optional)

---

## Backup Strategy

### Database Backup Script

```bash
#!/bin/bash
# /usr/local/bin/kabletown-backup.sh

BACKUP_DIR="/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "${BACKUP_DIR}"

# Backup database
mysqldump -h localhost -u jellyfin -p"${DB_PASSWORD}" \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  jellyfin | gzip > "${BACKUP_DIR}/backup_${DATE}.sql.gz"

# Retention: Keep 7 days
find "${BACKUP_DIR}" -name "backup_*.sql.gz" -mtime +7 -delete

# Copy to offsite (S3, GCS, etc.)
aws s3 cp "${BACKUP_DIR}/backup_${DATE}.sql.gz" s3://kabletown-backups/
```

- [ ] Backup script created
- [ ] Cron job scheduled: `0 2 * * * /usr/local/bin/kabletown-backup.sh`
- [ ] Offsite copy configured
- [ ] Retention policy set (7 days minimum)

### Backup Restore Test

```bash
# Test restore to temp database
mysql -h localhost -u root -p -e "CREATE DATABASE test_restore;"
zcat /backups/mysql/backup_20260313_020000.sql.gz | \
  mysql -h localhost -u root -p test_restore

# Verify data
mysql -h localhost -u root -p -e "SELECT COUNT(*) FROM test_restore.users;"
```

- [ ] Restore test successful
- [ ] Data integrity verified
- [ ] Production restore documented

---

## Security Hardening

### Database Security

```sql
-- Check for weak passwords
SELECT Username FROM users 
WHERE PasswordHash = (SELECT PasswordHash FROM users WHERE Username = 'admin');

-- Limit user permissions
REVOKE ALL PRIVILEGES ON jellyfin.* FROM 'jellyfin'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON jellyfin.* TO 'jellyfin'@'localhost';
FLUSH PRIVILEGES;
```

- [ ] Admin password changed
- [ ] Database user permissions minimized
- [ ] Remote MySQL access disabled

### Nginx Security

```nginx
# Rate limiting
limit_req_zone $binary_remote_addr zone=kabletown:10m rate=10r/s;

server {
    location /items {
        limit_req zone=kabletown burst=20 nodelay;
        # ... rest of config
    }
}
```

- [ ] Rate limiting configured
- [ ] Security headers set
- [ ] TLS 1.2+ enforced

### Application Security

- [ ] Input validation (query params, file uploads)
- [ ] Password hashing (SHA256 minimum, argon2 recommended)
- [ ] Token expiration policy (if implemented)
- [ ] CORS disabled (same-origin via Nginx)
- [ ] SQL injection prevention (parameterized queries)
- [ ] Path traversal prevention (output path validation)

---

## Post-Deployment Checklist

### First Boot

- [ ] All services running: `docker ps`
- [ ] No service crashes: `docker logs kabletown-*`
- [ ] Health checks pass: `curl https://yourdomain.com/health`

### User Testing

- [ ] Login works: Username + Password → AccessToken
- [ ] Device registration works: POST /Devices → AccessToken
- [ ] Item queries work: GET /Items → List of items
- [ ] Content playback works: HLS stream loads
- [ ] Live transcode works: Upload/queue transcode job

### Monitoring Setup

- [ ] Service uptime monitoring (UptimeRobot, Pingdom)
- [ ] Error rate monitoring (Sentry, Rollbar)
- [ ] Resource monitoring (Prometheus + Grafana, New Relic)
- [ ] Alert rules configured (disk space, CPU, memory)

### Documentation

- [ ] Admin documentation updated
- [ ] Runbook created (troubleshooting, restart procedures)
- [ ] Backup/restore procedures documented
- [ ] Contact info for support team

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker logs kabletown-auth-service

# Check MySQL
docker logs kabletown-mysql

# Restart failed service
docker-compose restart auth-service

# Check health
docker-compose ps
```

### Transcode Fails

```bash
# Check FFMPEG installation
which ffmpeg
ffmpeg -version

# Check transcoding directory
ls -la /transcode

# Manual FFmpeg test
ffmpeg -i /media/movie.mkv -c:v libx264 -f null -

# Check process limit
ulimit -n  # Should be > 1000
```

### Database Connection Issues

```bash
# Check MySQL
docker-compose exec mysql mysql -u jellyfin -p

# Check connection from service
docker-compose exec item-service \
  mysql -h mysql -u jellyfin -p jellyfin -e "SELECT 1"

# Check credentials in .env
cat .env | grep DB_
```

---

## Rollback Procedures

### Rollback to Previous Version

```bash
# Stop current services
docker-compose down

# Checkout previous git tag
cd /path/to/kabletown
git checkout v1.0.0

# Restart services
docker-compose pull
docker-compose up -d

# Verify services healthy
docker-compose ps
```

### Emergency Rollback (Database)

```bash
# Stop all services
docker-compose stop

# Restore database from backup
mysql -h localhost -u root -p < /backups/mysql/backup_20260312.sql

# Restart services
docker-compose start
```

---

## Sign-Off

- [ ] All pre-deployment tasks complete
- [ ] Database configured and optimized
- [ ] Docker Compose running
- [ ] Nginx proxy working
- [ ] Migrations applied
- [ ] Backup strategy in place
- [ ] Security hardened
- [ ] Documentation updated
- [ ] Team trained

**Deployed by:** ___________________  
**Date:** ___________________  
**Environment:** Production / Staging / Development
