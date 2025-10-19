# Pingo - Ping Statistics Monitor

A lightweight network monitoring tool that continuously pings a target host and visualizes latency statistics over time.

## Features

- **Continuous Monitoring**: Runs 30 pings every round and saves summary statistics
- **SQLite Storage**: Stores min/avg/max/stddev metrics with automatic 30-day retention
- **Web Dashboard**: Real-time charts accessible at `http://localhost:7777`
- **Interactive Visualization**:
  - Combined chart showing all metrics
  - Toggle visibility of individual metrics (min, avg, max, stddev)
  - Date range filtering for historical data analysis
  - Auto-refresh every 5 seconds
- **Configuration**: TOML config file support with CLI overrides
- **Cross-Platform**: Works on macOS, Linux (including Raspberry Pi)
- **Single Binary**: No external dependencies, embeds web UI

## Building

### Local Build
```bash
go build -o pingo main.go
```

### Cross-Compile for Raspberry Pi (64-bit)
```bash
GOOS=linux GOARCH=arm64 go build -o pingo-arm64 main.go
```

### Cross-Compile for Raspberry Pi (32-bit)
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o pingo-arm main.go
```

## Installation

### Quick Start
```bash
# Run with defaults
./pingo
```

### With Configuration File
```bash
# Create config directory
mkdir -p ~/.config/pingo

# Copy example config
cp config.toml.example ~/.config/pingo/config.toml

# Edit configuration as needed
nano ~/.config/pingo/config.toml

# Run
./pingo
```

## Configuration

### File Locations

Pingo follows XDG Base Directory specifications:

- **Config**: `~/.config/pingo/config.toml`
- **Data**: `~/.local/share/pingo/ping_stats.db`

### Default Values

| Setting | Default | Description |
|---------|---------|-------------|
| `port` | `7777` | Web server port |
| `target` | `8.8.8.8` | Host to ping |
| `ping_count` | `30` | Number of pings per round |
| `retention_days` | `30` | Days to retain data |
| `db_path` | `~/.local/share/pingo/ping_stats.db` | Database file path |

### Configuration Methods (in priority order)

1. **CLI flags** (highest priority)
2. **Config file** (`~/.config/pingo/config.toml`)
3. **Built-in defaults** (fallback)

### CLI Options

```bash
./pingo [options]

Options:
  -config string
        Path to config file (default "~/.config/pingo/config.toml")
  -db string
        Path to SQLite database file (overrides config)
  -pings int
        Number of pings per round (overrides config)
  -port string
        Web server port (overrides config)
  -retention int
        Number of days to retain ping data (overrides config)
  -target string
        Target host to ping (overrides config)
```

### Examples

```bash
# Use defaults
./pingo

# Custom port
./pingo -port 8080

# Monitor specific host with custom settings
./pingo -target 1.1.1.1 -pings 50 -retention 60

# Use custom config file
./pingo -config /etc/pingo/config.toml

# Override specific config values
./pingo -port 9000 -target 9.9.9.9
```

## Running as a Service on Raspberry Pi

See `pingo.service` for systemd service configuration.

## TODO

- [ ] Create `.deb` package for easy installation on Debian-based systems
- [ ] Create install script for automated service setup across different distros
