# bakdb Quick Start Guide

Get bakdb running in 5 minutes! 🚀

---

## 1️⃣ Install

### Linux/macOS (Recommended)

```bash
# Clone repo
git clone https://github.com/mtai0524/tui_backup_db.git
cd bakdb

# Run installer
chmod +x install.sh
./install.sh

# That's it! You can now run:
bakdb
```

### Windows

```bash
# Clone repo
git clone https://github.com/mtai0524/tui_backup_db.git
cd bakdb

# Build
go build -o bakdb.exe

# Run
bakdb.exe
```

### Or Build Locally (All Platforms)

```bash
cd bakdb
make build
./build/bakdb          # Linux/macOS
./build/bakdb.exe      # Windows
```

---

## 2️⃣ First Run

```bash
bakdb
```

You'll see:
```
┌─────────────────────────────────┐
│ Select Database Type            │
├─────────────────────────────────┤
│ ▸ MySQL                         │
│   PostgreSQL                    │
│   SQL Server                    │
└─────────────────────────────────┘
```

**Navigation:**
- `↑` `↓` arrow keys to select
- `Enter` to confirm

---

## 3️⃣ Enter Connection Details

```
Host            localhost
Port            3306
Username        root
Password        ••••••••
Database Name   my_database
Connection Str  (leave empty)
Binary Path     (leave empty)
Output Dir      (leave empty)
Backup Format   (SQL Server only)
```

**Tips:**
- `Tab` or `Enter` to move to next field
- Fields show defaults in light gray
- Connection String overrides individual fields

---

## 4️⃣ Run Backup

```
⟳ Backing up MySQL...

This may take a moment depending on the database size.
```

Wait for completion...

---

## 5️⃣ View Result

Success:
```
✔ Backup Successful!

File saved to: /home/user/my_database_20240514_143022.sql

(e) send via email  (r) restart  (q) quit
```

---

## 📧 Send via Email (Optional)

Press `e` to send:

```
✉ Send backup via Email

File: my_database_20240514_143022.sql

Gmail of you         you@gmail.com
App Password         ••••••••••••••••
Send to              recipient@example.com
Subject              Database Backup

         [ Send ]

Enter/Tab: next field  Ctrl+S: send  Esc: close
```

**Gmail Setup:**
1. Go to https://myaccount.google.com/apppasswords
2. Create App Password (16 characters)
3. Use that password here (not your Gmail password!)

---

## ⚙️ Configure Defaults (Optional)

Create `.env` file in project directory or `~/.bakdb/.env`:

```env
BAKDB_TYPE=MySQL
BAKDB_HOST=localhost
BAKDB_PORT=3306
BAKDB_USER=root
BAKDB_PASSWORD=secret
BAKDB_DATABASE=mydb
BAKDB_OUTPUT_DIR=~/backups
BAKDB_EMAIL_FROM=you@gmail.com
BAKDB_EMAIL_APP_PASSWORD=xxxx xxxx xxxx xxxx
BAKDB_EMAIL_TO=recipient@example.com
```

Next time you run bakdb, these fields will be pre-filled! ✨

---

## 🔧 Troubleshooting

### "Command not found: mysqldump"
```bash
# macOS
brew install mysql

# Ubuntu/Debian
sudo apt-get install mysql-client

# CentOS
sudo yum install mysql
```

### "Connection refused"
- Check database is running
- Check host, port, credentials are correct
- Firewall might be blocking connection

### Email not sending
- Check Gmail 2FA is enabled
- Use App Password (not main password)
- Verify email address is correct

---

## 📚 Next Steps

- 📖 Read [README.md](./README.md) for full documentation
- ⚙️ Check [Configuration](#-configuration) section
- 🐛 Report issues: https://github.com/mtai0524/tui_backup_db/issues

---

## 💡 Tips

- **First time?** Start with no `.env` file and enter details manually
- **Secure backups?** Use `.bak` format for SQL Server (native & faster)
- **Remote server?** Use `.sql` format for better portability
- **Large databases?** Run during off-peak hours
- **Important data?** Keep backups in multiple locations

Enjoy! 🎉
