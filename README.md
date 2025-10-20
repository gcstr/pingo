# Pingo - Ping Statistics Monitor

A lightweight network monitoring tool designed for Raspberry Pi that continuously pings a target host and visualizes latency statistics over time.

Optimized to run on resource-constrained devices like Raspberry Pi Zero, but also works on other platforms (Linux, macOS).

## Features

- **Continuous Monitoring**: Configurable ping rounds
- **SQLite Storage**: Stores metrics with configurable retention
- **Web Dashboard**: Real-time charts
- **Flexible Configuration**: TOML config file support with CLI overrides
- **Cross-Platform**: Optimized for Raspberry Pi, also runs on other Linux distributions and macOS
- **Single Binary**: No external dependencies, embeds web UI
- **Lightweight**: Minimal resource usage, perfect for Raspberry Pi Zero

## Installation

### Raspberry Pi / Debian-based Systems

Download and install the `.deb` package for automatic setup:

```bash
# Raspberry Pi (ARM64 - Raspberry Pi 3/4/5)
wget https://github.com/gcstr/pingo/releases/latest/download/pingo_0.1.0_linux_arm64.deb
sudo dpkg -i pingo_0.1.0_linux_arm64.deb

# Raspberry Pi (ARMv7 - Raspberry Pi Zero/2)
wget https://github.com/gcstr/pingo/releases/latest/download/pingo_0.1.0_linux_armv7.deb
sudo dpkg -i pingo_0.1.0_linux_armv7.deb

# x86_64 Linux
wget https://github.com/gcstr/pingo/releases/latest/download/pingo_0.1.0_linux_amd64.deb
sudo dpkg -i pingo_0.1.0_linux_amd64.deb
```

The `.deb` package will:
- Install the binary to `/usr/bin/pingo`
- Create the systemd service
- Enable and start the service automatically
- Dashboard will be available at `http://localhost:7777`

### RedHat/Fedora Systems

```bash
# ARM64
wget https://github.com/gcstr/pingo/releases/latest/download/pingo_0.1.0_linux_arm64.rpm
sudo rpm -i pingo_0.1.0_linux_arm64.rpm

# x86_64
wget https://github.com/gcstr/pingo/releases/latest/download/pingo_0.1.0_linux_amd64.rpm
sudo rpm -i pingo_0.1.0_linux_amd64.rpm
```

### Manual Installation (macOS / Other Linux)

Download the appropriate archive for your platform:

```bash
# Extract the archive
tar xzf pingo_0.1.0_*.tar.gz

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

## Building from Source

### Prerequisites

- Go 1.21 or later

### Local Build
```bash
go build -o pingo
```

### Cross-Compile for Raspberry Pi (64-bit)
```bash
GOOS=linux GOARCH=arm64 go build -o pingo-arm64
```

### Cross-Compile for Raspberry Pi (32-bit)
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -o pingo-arm
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
| `ping_count` | `5` | Number of pings per round |
| `retention_days` | `15` | Days to retain data |
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

## Managing the Service

If you installed via `.deb` or `.rpm` package, pingo runs as a systemd service:

```bash
# systemctl
sudo systemctl [status,stop,start,restart] pingo

# View logs
sudo journalctl -u pingo -f

# Disable auto-start on boot
sudo systemctl disable pingo

# Enable auto-start on boot
sudo systemctl enable pingo
```

For manual installation, see `pingo.service` for systemd service configuration.
