# bakdb Deployment Guide

Complete guide for deploying bakdb in production environments.

---

## 📦 Build & Package

### Single Platform Build

```bash
# Current platform
make build

# Output: ./build/bakdb (Linux/macOS) or ./build/bakdb.exe (Windows)
```

### Multi-Platform Release Build

```bash
make release

# Creates:
# - build/bakdb-linux-amd64
# - build/bakdb-linux-arm64
# - build/bakdb-macos-amd64
# - build/bakdb-macos-arm64
# - build/bakdb-windows-amd64.exe
```

---

## 🐧 Linux Deployment

### Ubuntu/Debian

#### 1. Install to System

```bash
# Build
make build

# Install
sudo make install

# Verify
bakdb --version
```

#### 2. Create Desktop Shortcut

```bash
sudo cp bakdb.desktop /usr/share/applications/
# Now available in application launcher
```

#### 3. As a Systemd Service

Create `/etc/systemd/system/bakdb.service`:

```ini
[Unit]
Description=bakdb Scheduled Backups
After=network.target

[Service]
Type=simple
User=backup
WorkingDirectory=/home/backup/.bakdb
EnvironmentFile=/home/backup/.bakdb/.env
ExecStart=/usr/local/bin/bakdb

[Install]
WantedBy=multi-user.target
```

Enable:
```bash
sudo systemctl daemon-reload
sudo systemctl enable bakdb
```

#### 4. Create Dedicated User (Recommended)

```bash
# Create backup user
sudo useradd -m -s /bin/bash backup

# Give permission to database tools
sudo usermod -aG mysql backup      # if needed
sudo usermod -aG postgres backup   # if needed

# Setup directories
sudo mkdir -p /var/backups/databases
sudo chown backup:backup /var/backups/databases
```

---

## 🍎 macOS Deployment

### Installation

```bash
# Build for specific architecture
make macos-arm64   # Apple Silicon
make macos-amd64   # Intel

# Copy to /usr/local/bin
sudo cp build/bakdb-macos-arm64 /usr/local/bin/bakdb
chmod +x /usr/local/bin/bakdb
```

### Launch Agent (Run at Login)

Create `~/Library/LaunchAgents/com.bakdb.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.bakdb</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/bakdb</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/bakdb.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/bakdb.log</string>
</dict>
</plist>
```

Load:
```bash
launchctl load ~/Library/LaunchAgents/com.bakdb.plist
```

---

## 🪟 Windows Deployment

### Standalone Executable

```bash
make windows

# Output: build/bakdb-windows-amd64.exe
```

Create batch file `bakdb.bat`:

```batch
@echo off
"C:\Program Files\bakdb\bakdb.exe" %*
```

### As Windows Service

Using NSSM (Non-Sucking Service Manager):

```batch
# Download from https://nssm.cc/download
nssm install bakdb "C:\Program Files\bakdb\bakdb.exe"
nssm set bakdb AppDirectory "C:\Program Files\bakdb"
nssm start bakdb
```

### Scheduled Task

```batch
schtasks /create /tn "bakdb-daily" ^
  /tr "C:\Program Files\bakdb\bakdb.exe" ^
  /sc daily /st 02:00
```

---

## 🐳 Docker Deployment

### Build Image

```bash
docker build -t bakdb:1.0.0 .
```

### Run Container

```bash
# Interactive mode
docker run -it \
  -v ~/.bakdb:/root/.bakdb \
  -v ~/backups:/backups \
  bakdb:1.0.0

# With environment file
docker run -it \
  --env-file ~/.bakdb/.env \
  -v ~/backups:/backups \
  bakdb:1.0.0
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  bakdb:
    image: bakdb:1.0.0
    container_name: bakdb
    environment:
      BAKDB_TYPE: MySQL
      BAKDB_HOST: mysql
      BAKDB_PORT: 3306
      BAKDB_USER: backup
      BAKDB_PASSWORD: ${DB_PASSWORD}
      BAKDB_DATABASE: mydb
      BAKDB_OUTPUT_DIR: /backups
    volumes:
      - ~/backups:/backups
      - ~/.bakdb/.env:/root/.bakdb/.env:ro
    networks:
      - internal
    depends_on:
      - mysql

  mysql:
    image: mysql:8.0
    container_name: mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    networks:
      - internal

networks:
  internal:
```

Run:
```bash
docker-compose up
```

---

## ☁️ Cloud Deployment

### AWS EC2

1. Launch instance (Ubuntu 20.04 LTS or later)
2. Connect via SSH
3. Run:

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y golang-go git build-essential
sudo apt-get install -y mysql-client postgresql-client

# Clone and build
git clone https://github.com/mtai0524/tui_backup_db.git
cd bakdb
chmod +x install.sh
./install.sh

# Configure
cp .env.example ~/.bakdb/.env
nano ~/.bakdb/.env
```

### Azure Container Instances

```bash
# Build and push to ACR
docker build -t bakdb:1.0.0 .
docker tag bakdb:1.0.0 myregistry.azurecr.io/bakdb:1.0.0
docker push myregistry.azurecr.io/bakdb:1.0.0

# Deploy
az container create \
  --resource-group mygroup \
  --name bakdb \
  --image myregistry.azurecr.io/bakdb:1.0.0 \
  --env-file /path/to/.env
```

### Google Cloud Run

```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/bakdb
gcloud run deploy bakdb \
  --image gcr.io/PROJECT_ID/bakdb \
  --memory 512Mi \
  --set-env-vars BAKDB_TYPE=MySQL
```

---

## 🔒 Security Best Practices

### 1. Protect `.env` File

```bash
chmod 600 ~/.bakdb/.env
sudo chown backup:backup ~/.bakdb/.env
```

### 2. Backup Directory Permissions

```bash
mkdir -p /var/backups/databases
chmod 700 /var/backups/databases
chown backup:backup /var/backups/databases
```

### 3. Use Secrets Management

**AWS Secrets Manager:**
```bash
aws secretsmanager create-secret --name bakdb-credentials \
  --secret-string '{"password":"xxx","appPassword":"yyy"}'
```

**Kubernetes Secrets:**
```bash
kubectl create secret generic bakdb-credentials \
  --from-file=.env
```

### 4. Audit Logging

```bash
# Log all bakdb operations
sudo journalctl -u bakdb -f
```

### 5. Network Security

- Run bakdb user with minimal privileges
- Use firewall to restrict database access
- Use VPN for remote database connections
- Enable TLS for all database connections

---

## 📊 Monitoring & Logging

### Syslog Integration

```bash
# Redirect bakdb output to syslog
bakdb 2>&1 | logger -t bakdb
```

### Log Rotation

Create `/etc/logrotate.d/bakdb`:

```
/var/log/bakdb.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 backup backup
    sharedscripts
    postrotate
        systemctl reload bakdb > /dev/null 2>&1 || true
    endscript
}
```

### Email Alerts

Configure `.env`:
```env
BAKDB_EMAIL_TO=admin@company.com
BAKDB_EMAIL_SUBJECT=Database Backup Report
```

---

## 🔄 Scheduled Backups

### cron Job (Linux)

```bash
# Edit crontab
crontab -e

# Daily at 2 AM
0 2 * * * /usr/local/bin/bakdb >> /var/log/bakdb.log 2>&1

# Multiple databases
0 2 * * * BAKDB_DATABASE=db1 /usr/local/bin/bakdb
0 3 * * * BAKDB_DATABASE=db2 /usr/local/bin/bakdb
```

### systemd Timer

Create `/etc/systemd/system/bakdb-daily.timer`:

```ini
[Unit]
Description=bakdb Daily Backup
Requires=bakdb-daily.service

[Timer]
OnCalendar=daily
OnCalendar=02:00

[Install]
WantedBy=timers.target
```

---

## 🧪 Testing Deployment

### Health Check

```bash
# Test binary
bakdb --version

# Test database connection
BAKDB_TYPE=MySQL bakdb

# Test email
BAKDB_EMAIL_FROM=test@gmail.com bakdb
```

### Backup Verification

```bash
# Check backup file
ls -lh ~/backups/
file ~/backups/*.sql

# Verify backup integrity (MySQL)
mysql -h localhost -u root -p < ~/backups/mydb_*.sql
```

---

## 📋 Checklist

- [ ] Built for target platform
- [ ] Database tools installed (`mysqldump`, `pg_dump`, `sqlcmd`)
- [ ] User account created (if running as service)
- [ ] Permissions set correctly
- [ ] `.env` file configured
- [ ] Backup directory created
- [ ] Tested backup process
- [ ] Tested email functionality
- [ ] Scheduled backups configured
- [ ] Monitoring and logging enabled
- [ ] Security hardening completed
- [ ] Documentation updated

---

## 📞 Support

For deployment issues:
- Check logs: `journalctl -u bakdb -f`
- Verify configuration: check `.env` file
- Test connectivity: `ping` database server
- Report issues: https://github.com/mtai0524/tui_backup_db/issues
