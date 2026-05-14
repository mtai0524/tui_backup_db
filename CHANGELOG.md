# Changelog

All notable changes to bakdb will be documented in this file.

## [1.0.0] - 2024-05-14

### Added
- ✨ Interactive TUI for database backup management
- 🗄️ Support for MySQL, PostgreSQL, and SQL Server
- 📧 Professional HTML email integration with:
  - Gradient styling and modern design
  - Database backup details (name, size, format, time)
  - Restore instructions for each backup format
  - Security warnings and best practices
  - Plain text fallback for compatibility
- 📝 Smart SQL Server format selection:
  - `.bak` files (native SQL Server backup)
  - `.sql` files (portable SQL scripts)
  - Auto-detection for local/remote servers
- ⚙️ Environment variable configuration (`.env`)
- 🔐 Secure credential handling
- 🎨 Real-time feedback with spinners
- 📦 Build system with Makefile
- 🚀 Installation scripts for easy setup
- 📖 Comprehensive documentation

### Features
- Interactive terminal UI with Bubble Tea
- Support for custom connection strings
- Optional path to database tools
- Configurable output directory
- Email modal with validation
- Tab navigation and keyboard shortcuts
- Result status display
- Error handling and user feedback

### Security
- No sensitive data logging
- Gmail App Password support (not plain passwords)
- Secure SMTP with TLS
- MIME multipart email structure

### Platform Support
- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

### Documentation
- Complete README with quick start guide
- Installation instructions for all platforms
- Configuration examples
- Email setup guide
- Troubleshooting section
- Development guide

---

## Planned for Future Releases

### v1.1.0
- [ ] Scheduled backups (cron support)
- [ ] Backup history and management UI
- [ ] Restore functionality
- [ ] Cloud storage support (S3, Google Cloud, Azure)
- [ ] Encryption for backup files
- [ ] Multi-database backup jobs

### v1.2.0
- [ ] Web UI interface
- [ ] REST API for integrations
- [ ] Webhook notifications
- [ ] Backup compression (.gz, .zip)
- [ ] Differential/incremental backups
- [ ] Custom backup pre/post scripts

### v2.0.0
- [ ] Rewrite in Rust for better performance
- [ ] Native mobile app
- [ ] Desktop applications (Electron)
- [ ] Agent-based backup system
- [ ] Advanced monitoring and analytics

---

## [Unreleased]

### In Development
- Scheduled backup feature
- Cloud storage integrations
- Backup encryption
