# cronlog

Lightweight wrapper for cron jobs that captures output, duration, and exit codes to a local SQLite store.

## Installation

```bash
go install github.com/yourusername/cronlog@latest
```

## Usage

Wrap any cron job command with `cronlog run`:

```bash
cronlog run -- /path/to/your/script.sh
```

Your crontab entry might look like:

```
0 2 * * * cronlog run -- /usr/local/bin/backup.sh
```

cronlog captures stdout, stderr, exit code, and duration, storing everything in a local SQLite database (`~/.cronlog/jobs.db` by default).

### Viewing logs

```bash
# List recent job runs
cronlog list

# Show output for a specific run
cronlog show <run-id>

# Tail the last N runs
cronlog list --last 10
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `~/.cronlog/jobs.db` | Path to SQLite database |
| `--job` | command name | Human-readable job label |
| `--timeout` | none | Kill job after duration |

### Example output

```
ID   JOB           STARTED              DURATION   EXIT
42   backup.sh     2024-11-01 02:00:03  1m24s      0
41   backup.sh     2024-10-31 02:00:01  1m31s      0
40   backup.sh     2024-10-30 02:00:02  0s         1
```

## Requirements

- Go 1.21+
- SQLite (via CGo or pure-Go driver)

## License

MIT © [yourusername](https://github.com/yourusername)