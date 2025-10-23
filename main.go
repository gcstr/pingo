package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Define CLI flags
	configPath := flag.String("config", getDefaultConfigPath(), "Path to config file")
	port := flag.String("port", "", "Web server port (overrides config)")
	retentionDays := flag.Int("retention", 0, "Number of days to retain ping data (overrides config)")
	pingCount := flag.Int("pings", 0, "Number of pings per round (overrides config)")
	target := flag.String("target", "", "Target host to ping (overrides config)")
	dbPath := flag.String("db", "", "Path to SQLite database file (overrides config)")

	flag.Parse()

	// Load config file
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// CLI flags override config file values
	if *port != "" {
		config.Port = *port
	}
	if *retentionDays > 0 {
		config.RetentionDays = *retentionDays
	}
	if *pingCount > 0 {
		config.PingCount = *pingCount
	}
	if *target != "" {
		config.Target = *target
	}
	if *dbPath != "" {
		config.DBPath = *dbPath
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory %s: %v", dbDir, err)
	}

	db, err := initDB(config.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Configuration: target=%s, pings=%d, retention=%d days, port=%s, db=%s",
		config.Target, config.PingCount, config.RetentionDays, config.Port, config.DBPath)

	// Run ping monitoring in background
	go runPingMonitor(db, config.Target, config.PingCount, config.RetentionDays)

	// Start web server (blocks)
	startWebServer(db, config.Port)
}
