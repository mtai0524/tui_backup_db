# bakdb - Enterprise Database Backup Manager

<div align="center">

![bakdb](https://img.shields.io/badge/bakdb-v1.0.0-blue)
![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

**A modern, interactive terminal-based database backup tool with professional HTML email support.**

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Usage](#-usage) • [Configuration](#-configuration)

</div>

---

## 🎯 Features

- **📦 Interactive TUI**: Modern terminal interface with Bubble Tea framework
- **🗄️ Multi-Database Support**: MySQL, PostgreSQL, SQL Server
- **📊 Backup Formats**:
  - `.sql` scripts (portable, works everywhere)
  - `.bak` files (SQL Server native format, compact & fast)
- **📧 Professional Email Integration**: 
  - HTML formatted emails with styling
  - Backup details, restore instructions, security warnings
  - Plain text fallback for compatibility
- **⚡ Smart Format Selection**: Auto-detect local/remote servers
- **🔐 Secure**: No sensitive data logging, supports App Passwords
- **🎨 Real-time Feedback**: Spinners and status messages
- **⚙️ Configurable**: `.env` file support for defaults

---

## 📋 Prerequisites

Ensure the following are installed:

- **Go** 1.18+ ([download](https://golang.org/dl/))
- **Database Tools** (depending on what you need):
  - `mysqldump` (MySQL/MariaDB) - [install guide](https://dev.mysql.com/doc/mysql-shell/8.0/en/mysql-shell-install.html)
  - `pg_dump` (PostgreSQL) - [install guide](https://www.postgresql.org/download/)
  - `sqlcmd` (SQL Server) - [install guide](https://docs.microsoft.com/en-us/sql/tools/sqlcmd-utility)

---

## 🚀 Installation

### Option 1: Quick Install (Recommended)

```bash
# Clone repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd bakdb

# Run installation script
chmod +x install.sh
./install.sh
```

This will:
- ✅ Check Go and Git
- ✅ Build the application
- ✅ Install to `/usr/local/bin`
- ✅ Create config directory `~/.bakdb`

### Option 2: Manual Build

```bash
# Clone repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd bakdb

# Build
make build
# or: go build -o bakdb

# Install (optional)
sudo cp bakdb /usr/local/bin/
```

### Option 3: Run Without Installing

```bash
# From project directory
./run.sh
# or
make dev
```

---

## ⚡ Quick Start

### 1. Run the application

```bash
bakdb
```

### 2. Select Database Type

```
┌─────────────────────────────────┐
│ Select Database Type            │
├─────────────────────────────────┤
│ ▸ MySQL                         │
│   PostgreSQL                    │
│   SQL Server                    │
└─────────────────────────────────┘
```

### 3. Enter Connection Details

```
┌─────────────────────────────────┐
│ Backup MySQL Database           │
├─────────────────────────────────┤
│ Host          │ localhost       │
│ Port          │ 3306            │
│ Username      │ root            │
│ Password      │ ••••••••        │
│ Database Name │ mydb            │
│ ...more fields...               │
│                                 │
│        [ Start Backup ]         │
└─────────────────────────────────┘
```

### 4. Backup Runs

```
⟳ Backing up MySQL...

This may take a moment depending on the database size.
```

### 5. View Result & Send Email

```
✔ Backup Successful!

File saved to: /home/user/mydb_20240514_143022.sql

(e) send via email  (r) restart  (q) quit
```

---

## 📖 Usage

### Via TUI

1. Press **↑↓** arrows to navigate
2. Press **Tab/Shift+Tab** to move between fields
3. Press **Enter** to confirm or move to next field
4. Press **Ctrl+C** or **Q** to quit
5. Press **E** (on result screen) to send backup via email

### SQL Server Format Selection

For SQL Server, you can choose:
- **`.bak`** (Native SQL Server backup)
  - Faster, more compact
  - Only works on SQL Server
  - Default for local servers
  
- **`.sql`** (SQL Script)
  - Portable, works anywhere
  - Larger file size
  - Default for remote servers

Set via UI or `.env`:
```env
BAKDB_BACKUP_FORMAT=.bak
```

---

## ⚙️ Configuration

### Environment Variables

Create `~/.bakdb/.env` or `./.env` with defaults:

```env
# Database Type: MySQL | PostgreSQL | SQL Server
BAKDB_TYPE=MySQL

# Connection (use ONE of these)
# Option A: Connection String
BAKDB_CONN_STRING=user:password@tcp(localhost:3306)/dbname

# Option B: Individual fields
BAKDB_HOST=localhost
BAKDB_PORT=3306
BAKDB_USER=root
BAKDB_PASSWORD=secret
BAKDB_DATABASE=mydb

# Optional
BAKDB_BINARY_PATH=/usr/bin/mysqldump
BAKDB_OUTPUT_DIR=~/backups
BAKDB_BACKUP_FORMAT=.bak

# Email Configuration
BAKDB_EMAIL_FROM=your-email@gmail.com
BAKDB_EMAIL_APP_PASSWORD=xxxx xxxx xxxx xxxx
BAKDB_EMAIL_TO=recipient@example.com
BAKDB_EMAIL_SUBJECT=Database Backup
```

### Email Setup (Gmail)

1. Enable 2-Factor Authentication on Gmail
2. Create an App Password: https://myaccount.google.com/apppasswords
3. Use the 16-character password in `BAKDB_EMAIL_APP_PASSWORD`
4. Configure `BAKDB_EMAIL_FROM` with your Gmail address

---

## 🏗️ Project Structure

```
bakdb/
├── main.go              # Entry point
├── ui/                  # TUI components
│   ├── model.go        # Bubble Tea model
│   ├── views.go        # View rendering
│   ├── updates.go      # Event handling
│   ├── email_modal.go  # Email UI
│   └── styles.go       # Styling
├── backup/              # Backup engine
│   └── engine.go       # Backup logic (MySQL, PostgreSQL, SQL Server)
├── email/               # Email utilities
│   └── email.go        # SMTP client, HTML/text formatting
├── config/              # Configuration
│   └── env.go          # .env parsing
├── Makefile            # Build automation
├── install.sh          # Installation script
├── run.sh              # Quick run script
└── go.mod/.env.example # Dependencies & template config
```

---

## 🔨 Development

### Build Commands

```bash
# Current platform
make build

# Specific platforms
make linux-amd64
make macos-amd64
make windows

# All platforms
make release

# Development (build + run)
make dev

# Install to system
sudo make install
sudo make uninstall

# Clean build artifacts
make clean
```

### Debug

```bash
# Show version info
bakdb --version

# Run with custom env file
BAKDB_ENV_FILE=/path/to/.env bakdb
```

---

## 📧 Email Features

### Professional HTML Emails

- Gradient header with emoji icons
- Database info cards (name, size, format, time)
- Restore instructions for each format
- Security warnings
- Professional footer

### Plain Text Fallback

Emails support both HTML and plain text for maximum compatibility.

---

## ⚠️ Security Notes

- **Passwords**: Never commit `.env` to version control
- **App Passwords**: Gmail App Passwords are security tokens, treat them like passwords
- **Backup Files**: Store backups securely, restrict access
- **Email**: Consider encrypting backups before emailing

---

## 🐛 Troubleshooting

### "Command not found: mysqldump"
- Install MySQL: `brew install mysql` (macOS) or `apt-get install mysql-client` (Linux)
- Ensure tool is in PATH

### Email authentication fails
- Check Gmail App Password (should be 16 chars with spaces)
- Verify 2FA is enabled on Gmail account
- Check sender email matches Gmail account

### .bak file not created locally
- Ensure SQL Server service is running
- Check directory permissions
- For remote servers, use `.sql` format instead

---

## 📝 License

MIT License - feel free to use commercially

---

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Commit changes
4. Push and create a Pull Request

---

## 📞 Support

- 📖 [Documentation](./README.md)
- 🐛 [Report Issues](https://github.com/mtai0524/tui_backup_db/issues)
- 💡 [Feature Requests](https://github.com/mtai0524/tui_backup_db/discussions)
