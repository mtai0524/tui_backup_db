# bakdb — Database Backup Manager

![bakdb](https://img.shields.io/badge/bakdb-v1.0.0-blue) ![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?logo=go) ![License](https://img.shields.io/badge/License-MIT-green)

An interactive terminal (TUI) tool to back up **MySQL**, **PostgreSQL**, and **SQL Server** databases, with optional Gmail delivery of the backup file.

## Features

- Interactive TUI (Bubble Tea), no flags to memorize
- MySQL, PostgreSQL, SQL Server
- SQL Server output as `.bak` (native) or `.sql` (portable script)
- Send backups via Gmail (HTML + plain-text email)
- Defaults from a `.env` file

## Requirements

- **Go** 1.18+
- The client tool for your database (bakdb shells out to it):

| Database   | Tool        | Linux (apt)                         | macOS (brew)               |
|------------|-------------|-------------------------------------|----------------------------|
| MySQL      | `mysqldump` | `apt install mysql-client`          | `brew install mysql-client`|
| PostgreSQL | `pg_dump`   | `apt install postgresql-client`     | `brew install postgresql`  |
| SQL Server | `sqlcmd`    | see below                           | `brew install sqlcmd`      |

**`sqlcmd` is required for SQL Server** (both `.bak` and `.sql`). If your OS has no package, install the standalone [go-sqlcmd](https://github.com/microsoft/go-sqlcmd) binary:

```bash
ARCH=$([ "$(uname -m)" = aarch64 ] && echo arm64 || echo amd64)
curl -fsSL -o /tmp/sqlcmd.tar.bz2 \
  "https://github.com/microsoft/go-sqlcmd/releases/latest/download/sqlcmd-linux-${ARCH}.tar.bz2"
tar -xjf /tmp/sqlcmd.tar.bz2 -C /tmp
install -m 755 /tmp/sqlcmd ~/.local/bin/sqlcmd   # ~/.local/bin must be in PATH
sqlcmd --version
```

## Install

```bash
git clone https://github.com/mtai0524/tui_backup_db.git
cd tui_backup_db

./install.sh          # build + install to /usr/local/bin (recommended)
# or: go build -o bakdb && ./bakdb
# or: ./run.sh         (build + run without installing)
# or: make release     (cross-platform binaries)
```

Docker:

```bash
docker build -t bakdb:1.0.0 .
docker run -it -v ~/.bakdb:/root/.bakdb -v ~/backups:/backups bakdb:1.0.0
```

## Usage

Run `bakdb`, then:

1. Pick the database type (↑↓, Enter).
2. Fill in connection details (Tab / Enter between fields; Database Name is required).
3. Select `[ Start Backup ]`.
4. On success, press `e` to email the file, `r` to restart, `q` to quit.

**SQL Server format** — choose in the UI or via `BAKDB_BACKUP_FORMAT`:

- `.bak` — native, faster/smaller, SQL Server only (default for local servers)
- `.sql` — portable script, works anywhere (default for remote servers)

## Configuration

Create `./.env` or `~/.bakdb/.env` (copy from [.env.example](.env.example)) to pre-fill defaults:

```env
BAKDB_TYPE=MySQL                 # MySQL | PostgreSQL | SQL Server
BAKDB_HOST=localhost
BAKDB_PORT=3306
BAKDB_USER=root
BAKDB_PASSWORD=secret
BAKDB_DATABASE=mydb
# BAKDB_CONN_STRING=...          # overrides the host/port/user/password fields
BAKDB_OUTPUT_DIR=~/backups
BAKDB_BACKUP_FORMAT=.bak         # SQL Server only

# Email (Gmail App Password — never commit .env)
BAKDB_EMAIL_FROM=you@gmail.com
BAKDB_EMAIL_APP_PASSWORD=xxxx xxxx xxxx xxxx
BAKDB_EMAIL_TO=recipient@example.com
```

**Gmail:** enable 2FA, create an [App Password](https://myaccount.google.com/apppasswords) (16 chars), and use it for `BAKDB_EMAIL_APP_PASSWORD` — not your normal password.

## Troubleshooting

- **`executable file not found: sqlcmd / mysqldump / pg_dump`** — install the client tool (see Requirements) and make sure it's on your `PATH`.
- **`i/o timeout` connecting to a hosted SQL Server** (e.g. databaseasp.net / MonsterASP) — use the **"Remote access for SSMS"** host, not "Local access". The local hostname resolves to an internal IP that's unreachable from outside.
- **`.bak` not created** — `.bak` writes on the SQL Server machine, so it only works for local servers. Use `.sql` for remote.
- **Email auth fails** — confirm 2FA is on and you're using a Gmail App Password (16 chars).

## Project layout

```
main.go      ui/        backup/engine.go   email/email.go   config/env.go
```

## License

MIT.
