# watcher-watch

A terminal UI for monitoring inotify file watcher usage per process on Linux.

## Installation

```bash
go install github.com/thinkbig1979/watcher-watch@latest
```

Or build from source:

```bash
git clone https://github.com/thinkbig1979/watcher-watch.git
cd watcher-watch
go build -o watcher-watch .
```

## Usage

```bash
./watcher-watch
```

### Controls

| Key | Action |
|-----|--------|
| `r` | Refresh |
| `q` / `Esc` | Quit |
| `Up` / `Down` | Navigate |

## Output

Displays a table of processes sorted by watch count:

- PID
- Process name
- Watch count
- Percentage of system limit

Usage color indicates severity: green (<50%), orange (50-80%), red (>80%).

## Requirements

- Linux (reads from `/proc`)
- Go 1.21+
