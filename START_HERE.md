# 🚀 START HERE - bakdb

Quick reference for installation and running bakdb.

---

## ⚡ 5-Minute Setup

### Linux/macOS

```bash
# Step 1: Clone repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Step 2: Run installer
chmod +x install.sh
./install.sh

# Step 3: Run application
bakdb
```

### Windows

```bash
# Step 1: Clone repository
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

# Step 2: Build application
go build -o bakdb.exe

# Step 3: Run application
./bakdb.exe
```

---

## ✅ Prerequisites Check

Before installation, make sure you have:

```bash
# Check Go
go version          # Should be 1.18+

# Check Git
git --version       # Any recent version

# Check database tools (at least one)
mysqldump --version    # For MySQL
pg_dump --version      # For PostgreSQL
sqlcmd -?              # For SQL Server
```

**Don't have these?**
→ See [INSTALL.md](./INSTALL.md) for detailed installation

---

## 📖 Documentation Map

| Document | Purpose |
|----------|---------|
| **START_HERE.md** | You are here! Quick reference |
| **INSTALL.md** | Detailed installation & troubleshooting |
| **README.md** | Features, configuration, usage |
| **QUICKSTART.md** | 5-minute hands-on guide |
| **DEPLOYMENT.md** | Production deployment |
| **IMPROVEMENTS.md** | Roadmap & future features |
| **CONTRIBUTING.md** | How to contribute |

---

## 🎯 Three Ways to Run

### Option 1: System-Wide Installation (Recommended)

```bash
./install.sh
bakdb  # Run from anywhere
```

**Pros:** Works from anywhere, professional setup
**Cons:** Requires permission to `/usr/local/bin`

### Option 2: Run From Project Directory

```bash
make build
./build/bakdb
```

**Pros:** No installation needed
**Cons:** Must be in project directory

### Option 3: Docker

```bash
docker build -t bakdb .
docker run -it bakdb
```

**Pros:** Isolated environment, no dependencies
**Cons:** Requires Docker installed

---

## 🔧 First Run - Quick Steps

1. **Start application**
   ```bash
   bakdb
   ```

2. **Select database type** (↑↓ arrows, Enter to select)
   - MySQL
   - PostgreSQL
   - SQL Server

3. **Fill connection details**
   - Host: `localhost`
   - Port: (auto-filled)
   - Username: `root` (or your user)
   - Password: (your password)
   - Database Name: (database to backup)

4. **Start backup** (navigate to button, press Enter)

5. **Wait** (spinning indicator shows progress)

6. **Send email** (optional: press 'e' to email)

7. **Done!** File saved to disk

---

## ⚙️ Configuration (Optional)

Pre-fill default values with `.env`:

```bash
# Copy template
cp .env.example .env

# Edit configuration
nano .env

# Or create in home directory
mkdir -p ~/.bakdb
cp .env.example ~/.bakdb/.env
```

**Example .env:**
```env
BAKDB_TYPE=MySQL
BAKDB_HOST=localhost
BAKDB_PORT=3306
BAKDB_USER=root
BAKDB_PASSWORD=secret
BAKDB_DATABASE=mydb
BAKDB_OUTPUT_DIR=~/backups
```

Next time you run bakdb, these fields will be pre-filled!

---

## 🔑 Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate menu |
| `←` `→` | Scroll fields |
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Enter` | Select / Submit |
| `Ctrl+C` | Quit application |
| `Q` | Quit (on result screen) |
| `R` | Restart backup |
| `E` | Send via email |

---

## 🐛 Common Issues

### "Command not found: bakdb"

```bash
# Solution 1: Use full path
/usr/local/bin/bakdb

# Solution 2: Add to PATH
export PATH="/usr/local/bin:$PATH"

# Solution 3: Reinstall
./install.sh
```

### "Cannot find mysqldump"

```bash
# Linux
sudo apt-get install mysql-client

# macOS
brew install mysql-client
```

### "Connection refused"

Check:
- [ ] Database is running
- [ ] Host is correct (`localhost` or IP)
- [ ] Port is correct (3306 for MySQL, 5432 for PostgreSQL)
- [ ] Credentials are correct
- [ ] Firewall allows connection

### Email not working?

See email setup in [INSTALL.md](./INSTALL.md#email-setup)

---

## 📦 Build Options

```bash
# Current platform
make build              # Linux/macOS/Windows

# Specific platform
make linux-amd64        # Linux
make macos-arm64        # macOS (Apple Silicon)
make windows            # Windows

# All platforms
make release            # Creates build/ directory

# Development mode
make dev                # Build + run
```

---

## 🧹 Cleanup

```bash
# Remove build artifacts
make clean

# Uninstall (if installed system-wide)
sudo rm /usr/local/bin/bakdb

# Remove config
rm -rf ~/.bakdb

# Remove project (keep backups!)
rm -rf tui_backup_db
```

---

## 📚 Next Steps

After first run:

1. **Read README.md** - Learn all features
2. **Setup .env** - Pre-fill defaults (optional)
3. **Schedule backups** - Setup daily/weekly backups (see DEPLOYMENT.md)
4. **Configure email** - Setup automatic email (see INSTALL.md)
5. **Explore features** - Try all options in the UI

---

## 🎯 Production Setup

For automated backups:

```bash
# Linux cron job (daily at 2 AM)
crontab -e
# Add: 0 2 * * * /usr/local/bin/bakdb

# Or create Systemd timer
# See DEPLOYMENT.md for details

# Windows Task Scheduler
# See DEPLOYMENT.md for details
```

---

## 📞 Getting Help

1. **Check documentation**
   - [INSTALL.md](./INSTALL.md) - Installation issues
   - [README.md](./README.md) - Feature usage
   - [QUICKSTART.md](./QUICKSTART.md) - Quick examples

2. **Search issues**
   - https://github.com/mtai0524/tui_backup_db/issues

3. **Open new issue**
   - Describe problem clearly
   - Include OS, Go version, database type
   - Provide error messages

4. **Contributing**
   - See [CONTRIBUTING.md](./CONTRIBUTING.md)
   - All help is welcome!

---

## 🎉 Success Indicators

You know it's working when:

- ✅ Application starts without errors
- ✅ Database connection succeeds
- ✅ Backup file is created
- ✅ Backup contains data (check file size)
- ✅ Email sends successfully (if configured)

---

## 💡 Pro Tips

- **First time?** Leave most fields empty, use defaults
- **Slow backup?** Check database size and network
- **Large file?** Use `.bak` format for SQL Server (faster)
- **Email failing?** Use Gmail App Password, not main password
- **Need help?** Check error message carefully, google the error

---

## 🚀 Ready to Start?

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db
chmod +x install.sh
./install.sh
bakdb
```

Enjoy! 🎊

---

## 📋 Checklist

- [ ] Cloned repository
- [ ] Installed prerequisites (Go, Git, database tools)
- [ ] Ran installation script OR manual build
- [ ] Successfully ran bakdb
- [ ] Created first backup
- [ ] Read README.md
- [ ] (Optional) Setup .env file
- [ ] (Optional) Setup email
- [ ] (Optional) Schedule automated backups

---

## 🔗 Useful Links

- 🐙 **GitHub:** https://github.com/mtai0524/tui_backup_db
- 📖 **Documentation:** See links above
- 🐛 **Report Issues:** https://github.com/mtai0524/tui_backup_db/issues
- 💬 **Discussions:** https://github.com/mtai0524/tui_backup_db/discussions

---

Happy backing up! 🎈
