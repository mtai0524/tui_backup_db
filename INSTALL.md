# 📦 Installation & Setup Guide - bakdb

Complete step-by-step guide to install and run bakdb.

---

## ⚡ Quick Start (5 minutes)

### For Linux/macOS Users

```bash
# 1. Clone the repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# 2. Run installer (automatic setup)
chmod +x install.sh
./install.sh

# 3. Done! Run bakdb from anywhere:
bakdb
```

### For Windows Users

```bash
# 1. Clone repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# 2. Build
go build -o bakdb.exe

# 3. Run
./bakdb.exe
```

---

## 📋 Prerequisites

Make sure you have these installed:

### 1. Go Programming Language

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install -y golang-go
go version  # Verify (should be 1.18+)
```

**macOS:**
```bash
brew install go
go version  # Verify
```

**Windows:**
- Download from https://golang.org/dl/
- Run installer
- Verify: `go version` in Command Prompt

### 2. Git

**Linux:**
```bash
sudo apt-get install -y git
```

**macOS:**
```bash
brew install git
```

**Windows:**
- Download from https://git-scm.com/
- Run installer

### 3. Database Tools (Choose What You Need)

#### MySQL/MariaDB

**Linux:**
```bash
sudo apt-get install -y mysql-client
```

**macOS:**
```bash
brew install mysql-client
```

**Windows:**
- Download from https://dev.mysql.com/downloads/mysql/
- Or use: `choco install mysql` (if Chocolatey installed)

#### PostgreSQL

**Linux:**
```bash
sudo apt-get install -y postgresql-client
```

**macOS:**
```bash
brew install postgresql
```

**Windows:**
- Download from https://www.postgresql.org/download/
- Or use: `choco install postgresql` (if Chocolatey installed)

#### SQL Server (sqlcmd)

**Linux:**
```bash
# Ubuntu/Debian
curl https://packages.microsoft.com/keys/microsoft.asc | sudo apt-key add -
sudo add-apt-repository "$(curl https://packages.microsoft.com/config/ubuntu/$(lsb_release -rs)/mssql-server.list)"
sudo apt-get install -y mssql-tools
```

**macOS:**
```bash
brew tap microsoft/mssql-release https://github.com/Microsoft/homebrew-mssql-release
brew install mssql-tools
```

**Windows:**
- Download from https://learn.microsoft.com/en-us/sql/tools/sqlcmd/sqlcmd-utility
- Or use: `choco install sqlserver-cmdlineutils`

---

## 🚀 Installation Methods

### Method 1: Automatic Installation (Recommended)

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db
chmod +x install.sh
./install.sh
```

**What it does:**
- ✅ Checks Go and Git installation
- ✅ Builds the application
- ✅ Installs to `/usr/local/bin`
- ✅ Creates `~/.bakdb` config directory
- ✅ Copies `.env.example` template

**After installation:**
```bash
bakdb  # Run from anywhere!
```

### Method 2: Manual Build & Run

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Build
go build -o bakdb

# Run from current directory
./bakdb
```

### Method 3: Using Makefile

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Build
make build

# Run
./build/bakdb

# Or install system-wide
sudo make install

# Then run from anywhere
bakdb
```

### Method 4: Quick Run Script

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Auto-builds and runs
./run.sh
```

### Method 5: Docker

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Build image
docker build -t bakdb:1.0.0 .

# Run
docker run -it \
  -v ~/.bakdb:/root/.bakdb \
  -v ~/backups:/backups \
  bakdb:1.0.0
```

---

## 🎯 Running bakdb

### After Installation

Simply type in terminal:
```bash
bakdb
```

### First Run

1. **Select Database Type**
   ```
   ┌─────────────────────────────────┐
   │ Select Database Type            │
   ├─────────────────────────────────┤
   │ ▸ MySQL                         │
   │   PostgreSQL                    │
   │   SQL Server                    │
   └─────────────────────────────────┘
   ```
   - Use `↑` `↓` arrow keys
   - Press `Enter` to select

2. **Enter Connection Details**
   ```
   Host:           localhost
   Port:           3306
   Username:       root
   Password:       ••••••
   Database Name:  mydb
   ```
   - Use `Tab` to move between fields
   - Fields are optional (except Database Name)

3. **Start Backup**
   - Navigate to `[ Start Backup ]` button
   - Press `Enter`

4. **Wait for Completion**
   ```
   ⟳ Backing up MySQL...
   This may take a moment depending on the database size.
   ```

5. **View Result**
   ```
   ✔ Backup Successful!
   
   File saved to: /home/user/mydb_20240514_143022.sql
   
   (e) send via email  (r) restart  (q) quit
   ```

---

## ⚙️ Configuration (Optional)

Create a `.env` file to pre-fill default values:

### Create Config File

```bash
# Option 1: In project directory
cp .env.example .env
nano .env

# Option 2: In home directory
mkdir -p ~/.bakdb
cp .env.example ~/.bakdb/.env
nano ~/.bakdb/.env
```

### Example Configuration

```env
# Database Type
BAKDB_TYPE=MySQL

# Connection Settings
BAKDB_HOST=localhost
BAKDB_PORT=3306
BAKDB_USER=root
BAKDB_PASSWORD=your_password
BAKDB_DATABASE=mydb

# Optional
BAKDB_OUTPUT_DIR=~/backups
BAKDB_BACKUP_FORMAT=.bak

# Email (for Gmail)
BAKDB_EMAIL_FROM=your-email@gmail.com
BAKDB_EMAIL_APP_PASSWORD=xxxx xxxx xxxx xxxx
BAKDB_EMAIL_TO=recipient@example.com
BAKDB_EMAIL_SUBJECT=Database Backup
```

### Email Setup

1. **Enable 2-Factor Authentication**
   - Go to https://myaccount.google.com/security

2. **Create App Password**
   - Go to https://myaccount.google.com/apppasswords
   - Select "Mail" and "Windows Computer"
   - Copy the 16-character password

3. **Set in .env**
   ```env
   BAKDB_EMAIL_FROM=your-email@gmail.com
   BAKDB_EMAIL_APP_PASSWORD=xxxx xxxx xxxx xxxx
   ```

---

## ✅ Verification

### Test Installation

```bash
# Check if bakdb is installed
which bakdb

# Show version
bakdb --version

# Test database connection (without backup)
bakdb
```

### Test Email (Optional)

1. Run bakdb
2. Complete a backup
3. Press `e` to send via email
4. Check recipient's inbox

---

## 🐛 Troubleshooting

### "Command not found: bakdb"

**After `make install`?**
```bash
# Add to PATH
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Or use full path:**
```bash
/usr/local/bin/bakdb
```

### "Cannot find mysqldump"

**Install MySQL client:**
```bash
# Linux
sudo apt-get install mysql-client

# macOS
brew install mysql-client

# Windows - Download from:
# https://dev.mysql.com/downloads/mysql/
```

### "Cannot find pg_dump"

**Install PostgreSQL client:**
```bash
# Linux
sudo apt-get install postgresql-client

# macOS
brew install postgresql
```

### "Cannot find sqlcmd"

**Install SQL Server tools:**
```bash
# Linux
sudo apt-get install mssql-tools

# macOS
brew install mssql-tools
```

### "Connection refused"

Check:
- [ ] Database is running: `ping <host>`
- [ ] Port is correct
- [ ] Username & password are correct
- [ ] Firewall allows connection

### "Email not sending"

Check:
- [ ] Gmail 2FA is enabled
- [ ] Using App Password (not main password)
- [ ] Email address is correct
- [ ] SMTP access allowed in Gmail security settings

### Build Error: "Go not found"

```bash
# Install Go from:
# https://golang.org/dl/

# Verify:
go version
```

---

## 📚 Documentation

After installation, read:
- **QUICKSTART.md** - 5-minute quick start
- **README.md** - Complete features & usage
- **DEPLOYMENT.md** - Production deployment

---

## 🔄 Uninstall

### If Installed System-Wide

```bash
# Remove binary
sudo rm /usr/local/bin/bakdb

# Remove config (optional)
rm -rf ~/.bakdb
```

### If Running Locally

Just delete the project directory:
```bash
rm -rf tui_backup_db
```

---

## 🆘 Need Help?

- 📖 Check **README.md** for features
- ⚡ Check **QUICKSTART.md** for quick examples
- 🔨 Check **Makefile** for build options
- 🐛 Report issues: https://github.com/mtai0524/tui_backup_db/issues

---

## 🎉 You're Ready!

After installation, you can:

```bash
# Run backups
bakdb

# Schedule daily backups
crontab -e
# Add: 0 2 * * * /usr/local/bin/bakdb

# View help
make help

# Build for other platforms
make release
```

Enjoy! 🚀
