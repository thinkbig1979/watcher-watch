# watcher-watch

[![Built with Claude](https://img.shields.io/badge/Built%20with-Claude-blue)](https://claude.ai)

A terminal UI for monitoring inotify file watcher usage per process on Linux.

## Installation

### Download Binary

Download the latest release from the [Releases page](https://github.com/thinkbig1979/watcher-watch/releases).

```bash
# Example for Linux amd64
curl -LO https://github.com/thinkbig1979/watcher-watch/releases/latest/download/watcher-watch_linux_amd64.tar.gz
tar xzf watcher-watch_linux_amd64.tar.gz
sudo mv watcher-watch /usr/local/bin/
```

### Go Install

```bash
go install github.com/thinkbig1979/watcher-watch@latest
```

### Build from Source

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
