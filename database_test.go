package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestInitDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='ping_stats'").Scan(&tableName)
	if err != nil {
		t.Fatalf("ping_stats table not found: %v", err)
	}
	if tableName != "ping_stats" {
		t.Errorf("Expected table name 'ping_stats', got %s", tableName)
	}
}

func TestSavePingStats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	stats := &PingStats{
		Timestamp: time.Now(),
		Min:       10.5,
		Avg:       12.3,
		Max:       15.7,
		StdDev:    2.1,
	}

	err = savePingStats(db, stats, 30)
	if err != nil {
		t.Fatalf("Failed to save ping stats: %v", err)
	}

	// Verify data was saved
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM ping_stats").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 record, got %d", count)
	}
}

func TestGetRecentStats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Insert test data
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		stats := &PingStats{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Min:       10.0 + float64(i),
			Avg:       12.0 + float64(i),
			Max:       15.0 + float64(i),
			StdDev:    2.0,
		}
		err = savePingStats(db, stats, 30)
		if err != nil {
			t.Fatalf("Failed to save test data: %v", err)
		}
	}

	// Get recent stats
	stats, err := getRecentStats(db, 3)
	if err != nil {
		t.Fatalf("Failed to get recent stats: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 stats, got %d", len(stats))
	}

	// Verify chronological order (oldest first)
	if stats[0].Min > stats[1].Min {
		t.Error("Stats are not in chronological order")
	}
}

func TestGetStatsByDateRange(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Insert test data with specific timestamps
	baseTime := time.Date(2025, 10, 19, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		stats := &PingStats{
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Min:       10.0 + float64(i),
			Avg:       12.0 + float64(i),
			Max:       15.0 + float64(i),
			StdDev:    2.0,
		}
		err = savePingStats(db, stats, 30)
		if err != nil {
			t.Fatalf("Failed to save test data: %v", err)
		}
	}

	// Query for middle 3 hours
	startDate := baseTime.Add(1 * time.Hour).Format("2006-01-02T15:04:05")
	endDate := baseTime.Add(3 * time.Hour).Format("2006-01-02T15:04:05")

	stats, err := getStatsByDateRange(db, startDate, endDate)
	if err != nil {
		t.Fatalf("Failed to get stats by date range: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 stats in range, got %d", len(stats))
	}

	// Verify first entry
	if stats[0].Min != 11.0 {
		t.Errorf("Expected first min 11.0, got %f", stats[0].Min)
	}
}

func TestRetentionPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Insert old data (40 days ago)
	oldTime := time.Now().AddDate(0, 0, -40)
	oldStats := &PingStats{
		Timestamp: oldTime,
		Min:       10.0,
		Avg:       12.0,
		Max:       15.0,
		StdDev:    2.0,
	}
	err = savePingStats(db, oldStats, 30)
	if err != nil {
		t.Fatalf("Failed to save old data: %v", err)
	}

	// Insert recent data
	recentStats := &PingStats{
		Timestamp: time.Now(),
		Min:       11.0,
		Avg:       13.0,
		Max:       16.0,
		StdDev:    2.5,
	}
	err = savePingStats(db, recentStats, 30)
	if err != nil {
		t.Fatalf("Failed to save recent data: %v", err)
	}

	// Verify only recent data remains
	stats, err := getRecentStats(db, 10)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if len(stats) != 1 {
		t.Errorf("Expected 1 stat after retention cleanup, got %d", len(stats))
	}

	if len(stats) > 0 && stats[0].Min != 11.0 {
		t.Errorf("Expected recent data to remain, got min %f", stats[0].Min)
	}
}

func TestGetStatsByDateRangeInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := initDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	_, err = getStatsByDateRange(db, "invalid-date", "2025-10-19T15:00:00")
	if err == nil {
		t.Error("Expected error for invalid start date format, got nil")
	}

	_, err = getStatsByDateRange(db, "2025-10-19T15:00:00", "invalid-date")
	if err == nil {
		t.Error("Expected error for invalid end date format, got nil")
	}
}
