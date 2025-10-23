package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Port          string `toml:"port"`
	Target        string `toml:"target"`
	PingCount     int    `toml:"ping_count"`
	RetentionDays int    `toml:"retention_days"`
	DBPath        string `toml:"db_path"`
}

func getDefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: Could not get user home directory: %v\n\n", err)
		fmt.Fprintln(os.Stderr, "Pingo requires $HOME to be set to follow XDG Base Directory specifications.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Please set the HOME environment variable before running pingo:")
		fmt.Fprintln(os.Stderr, "  export HOME=/home/yourusername")
		fmt.Fprintln(os.Stderr, "  pingo")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or if installing as root/with sudo:")
		fmt.Fprintln(os.Stderr, "  sudo -E pingo    # Preserves your HOME")
		fmt.Fprintln(os.Stderr, "  # or")
		fmt.Fprintln(os.Stderr, "  sudo HOME=/root pingo")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "For systemd services, add this to your service file:")
		fmt.Fprintln(os.Stderr, "  [Service]")
		fmt.Fprintln(os.Stderr, "  Environment=\"HOME=/home/yourusername\"")
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}
	return filepath.Join(home, ".local", "share", "pingo")
}

func getDefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: Could not get user home directory: %v\n\n", err)
		fmt.Fprintln(os.Stderr, "Pingo requires $HOME to be set to follow XDG Base Directory specifications.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Please set the HOME environment variable before running pingo:")
		fmt.Fprintln(os.Stderr, "  export HOME=/home/yourusername")
		fmt.Fprintln(os.Stderr, "  pingo")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Or if installing as root/with sudo:")
		fmt.Fprintln(os.Stderr, "  sudo -E pingo    # Preserves your HOME")
		fmt.Fprintln(os.Stderr, "  # or")
		fmt.Fprintln(os.Stderr, "  sudo HOME=/root pingo")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "For systemd services, add this to your service file:")
		fmt.Fprintln(os.Stderr, "  [Service]")
		fmt.Fprintln(os.Stderr, "  Environment=\"HOME=/home/yourusername\"")
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
		return "" // unreachable
	}
	return filepath.Join(home, ".config", "pingo")
}

func getDefaultConfigPath() string {
	return filepath.Join(getDefaultConfigDir(), "config.toml")
}

func getDefaultDBPath() string {
	return filepath.Join(getDefaultDataDir(), "ping_stats.db")
}

func getDefaultConfig() Config {
    return Config{
        Port:          "7777",
        Target:        "8.8.8.8",
        PingCount:     5,
        RetentionDays: 15,
        DBPath:        getDefaultDBPath(),
    }
}

func loadConfig(configPath string) (Config, error) {
	config := getDefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, using defaults", configPath)
		return config, nil
	}

	// Read and parse config file
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	log.Printf("Loaded config from %s", configPath)
	return config, nil
}
