# bakdb GUI Desktop — Design

**Date:** 2026-06-13
**Status:** Approved (pending spec review)

## Goal

Convert bakdb from a terminal UI (TUI) into a cross-platform **desktop GUI application** with its own window — buttons, forms, mouse interaction — so users run it as a normal app instead of in a terminal. Reuse the existing Go backup/email/config engine unchanged.

## Constraints & Decisions

- **Cross-platform**, Go-based GUI using **Fyne** (compiles to one native binary per OS: Linux / Windows / macOS).
- **Reuse the existing engine unchanged**: `backup/`, `email/`, `config/` are untouched. Only the presentation layer changes.
- **Feature parity with the TUI** plus: auto-fill the form from `.env` on startup, remember the last output dir.
- **Sidebar + content layout** with 4 sections: Backup, Email, History, Settings.
- Build dependencies: Fyne needs system libs on Linux (`libgl1-mesa-dev`, `xorg-dev` / equivalent). Accepted.

## Architecture

```
main.go            → gui.Run()                  (changed: drop Bubble Tea, launch Fyne)
gui/               → NEW: Fyne desktop app
  app.go           → main window, sidebar nav (4 items), routes to views
  state.go         → AppState: loaded config.Defaults + last output dir; helpers
  view_backup.go   → DB-type selector + connection form + run + result
  view_email.go    → Gmail config + attachment picker + send
  view_history.go  → list backup files in output dir (newest first), per-row Send/Open
  view_settings.go → default output dir, optional binary path, Save to .env
backup/ email/ config/   → UNCHANGED — gui calls into them
ui/                → DELETED after the GUI is verified working
```

The GUI is a thin layer: each view collects input, builds a `backup.Config` or
`email.Config` struct, and calls the existing public functions.

### Engine functions consumed (already exist, no changes)

- `config.Load() config.Defaults` — read `.env` / env vars (priority order already implemented)
- `config.NormalizeType(string) string` — canonicalize DB type
- `backup.ExecuteBackup(backup.Config) (string, error)` — run backup, returns output file path
- `email.Send(email.Config, attachmentPath string) error` — send Gmail with attachment

## Components / Views

### Main window (app.go)
- Window ~900×600, resizable.
- Left **sidebar** with 4 nav items: 🗄 Backup · ✉ Email · 🕘 History · ⚙ Settings.
- Right **content area** swaps based on selection. Opens on **Backup** by default.
- Sidebar selection updates the content via a simple router (set content container's object).

### AppState (state.go)
- Holds `config.Defaults` loaded once at startup.
- Tracks `lastOutputDir` and `lastBackupFile` (the most recent successful backup path) so the Email view can pre-fill the attachment and History can default its directory.
- Provides the resolved default output dir for History and Settings.

### View: Backup (view_backup.go)
- **DB type**: radio group (MySQL / PostgreSQL / SQL Server), default from `Defaults.Type` or MySQL.
- **Connection form**: Host, Port, User, Password (masked), Database (required), Output dir — all pre-filled from `Defaults`. Optional Connection String field (overrides host/port/user/password, matching engine behavior).
- **Format** (SQL Server only): `.bak` / `.sql` radio, shown only when SQL Server is selected; from `Defaults.BackupFormat`.
- **Start Backup** button → validates Database is non-empty → runs `backup.ExecuteBackup` on a **background goroutine** (UI stays responsive, shows a progress/spinner state) → on success shows the file path with **[Open folder]** and **[Send Email]** buttons; on error shows the error message.
- **[Send Email]** switches to the Email view with the attachment pre-filled.
- **[Open folder]** opens the containing directory in the OS file manager.

### View: Email (view_email.go)
- Fields: Gmail from, App Password (masked), To (comma/semicolon separated), Subject — pre-filled from `Defaults`.
- **Attachment**: defaults to the last successful backup file (from AppState); a file picker lets the user choose another.
- **Send** button → builds `email.Config` (split To into `[]string`) → calls `email.Send` on a background goroutine → shows success or error.

### View: History (view_history.go)
- Lists files in the resolved output dir matching backup extensions (`.sql`, `.bak`, and compressed variants), sorted newest → oldest, showing name + size.
- Per row: **[Send]** (jump to Email with that file as attachment) and **[Open]** (open in file manager).
- Empty state if directory missing or no files.

### View: Settings (view_settings.go)
- Default output dir, optional binary path (sqlcmd/mysqldump/pg_dump).
- **Save to .env** writes a `.env` file in the working dir (or `~/.bakdb/.env`) with the current Backup + Email + Settings values, so next launch auto-fills. Never logs secrets to stdout.

## Data Flow

1. **Startup**: `main.go` → `gui.Run()` → `config.Load()` into `AppState` → build window → show Backup view with fields pre-filled.
2. **Backup**: user edits form → Start → goroutine runs `ExecuteBackup` → result posted back to UI thread (via Fyne's `fyne.Do`/channel) → AppState stores `lastBackupFile`.
3. **Email**: from result or History → Email view pre-filled → Send → goroutine `email.Send` → result shown.
4. **Settings save**: write `.env`, reload `Defaults` into AppState.

## Error Handling

- All long-running calls (backup, email) run off the UI goroutine; results marshalled back to the UI thread so the window never freezes.
- Engine errors (e.g. `sqlcmd` not found, connection timeout, auth failure) are shown verbatim in a result/error label or dialog — the engine already returns descriptive messages.
- Validation before run: Database name required for backup; From/AppPassword/To required for email. Show inline messages, don't crash.
- Missing output dir in History → friendly empty state, not an error.

## Testing

- **Engine** is already covered and unchanged — no new engine tests.
- **GUI logic** that is pure (no widgets) is unit-tested: To-string → `[]string` splitting, building `backup.Config`/`email.Config` from form values, output-dir file listing/sorting, `.env` serialization.
- **Manual verification**: build on Linux, launch the window, run a backup end-to-end, send an email, browse history, save settings. (TUI removal happens only after this passes.)

## Out of Scope (YAGNI)

- Scheduling/cron, multiple saved connection profiles, themes, in-app DB browser, auto-update. Can come later; not in this version.

## Migration / Packaging

- After the GUI is verified, delete `ui/` (TUI) and update `main.go`.
- Update `Makefile` cross-compile targets for Fyne (Fyne cross-compiling needs CGo / `fyne-cross`; document this). `make package` archives stay.
- Update README: it's now a desktop app; note Linux build-time system libs.
- `bakdb.desktop` updated: `Terminal=false` (it's a windowed app now).
```
