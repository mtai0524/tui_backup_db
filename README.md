# bakdb

`bakdb` is a modern, terminal-based database backup tool built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

## Features

- **Interactive TUI**: Sleek interface for selecting databases and entering credentials.
- **Support for MySQL, PostgreSQL & SQL Server**: Uses native tools (`mysqldump`, `pg_dump`, and `sqlcmd`) for reliable backups.
- **Real-time Feedback**: Spinner and status messages during execution.
- **Automatic Port Detection**: Defaults ports based on database selection.

## Prerequisites

Ensure you have the following installed and in your PATH:
- [Go](https://golang.org/dl/) (to build/run)
- `mysqldump` (for MySQL backups)
- `pg_dump` (for PostgreSQL backups)
- `sqlcmd` (for SQL Server backups)

## How to Run

1. Clone or copy the project files.
2. Initialize and install dependencies:
   ```bash
   go mod tidy
   ```
3. Run the application:
   ```bash
   go run main.go
   ```

## Usage

1. **Select Database**: Use the arrow keys and Enter to select between MySQL and PostgreSQL.
2. **Enter Credentials**: Fill in the Host, Port, User, Password, and Database Name. Use `Tab` to navigate between fields.
3. **Start Backup**: Navigate to the "Start Backup" button and press `Enter`.
4. **View Result**: Once complete, the tool will display the success message and the path to the generated `.sql` file.

## Project Structure

- `main.go`: Entry point.
- `ui/`: TUI components and logic.
- `backup/`: Core backup execution logic using `os/exec`.
