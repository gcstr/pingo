package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
    config := getDefaultConfig()

	if config.Port != "7777" {
		t.Errorf("Expected default port 7777, got %s", config.Port)
	}
	if config.Target != "8.8.8.8" {
		t.Errorf("Expected default target 8.8.8.8, got %s", config.Target)
	}
    if config.PingCount != 5 {
        t.Errorf("Expected default ping count 5, got %d", config.PingCount)
    }
    if config.RetentionDays != 15 {
        t.Errorf("Expected default retention days 15, got %d", config.RetentionDays)
    }
}

func TestLoadConfigNonExistent(t *testing.T) {
	config, err := loadConfig("/nonexistent/config.toml")
	if err != nil {
		t.Errorf("Expected no error for non-existent config, got %v", err)
	}

	// Should return default config
	if config.Port != "7777" {
		t.Errorf("Expected default port when config doesn't exist, got %s", config.Port)
	}
}

func TestLoadConfigValid(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
port = "8080"
target = "1.1.1.1"
ping_count = 50
retention_days = 60
db_path = "/tmp/test.db"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", config.Port)
	}
	if config.Target != "1.1.1.1" {
		t.Errorf("Expected target 1.1.1.1, got %s", config.Target)
	}
	if config.PingCount != 50 {
		t.Errorf("Expected ping count 50, got %d", config.PingCount)
	}
	if config.RetentionDays != 60 {
		t.Errorf("Expected retention days 60, got %d", config.RetentionDays)
	}
	if config.DBPath != "/tmp/test.db" {
		t.Errorf("Expected db path /tmp/test.db, got %s", config.DBPath)
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	// Create temporary invalid config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	invalidContent := `
port = "8080
invalid toml syntax
`

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err = loadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid config file, got nil")
	}
}
