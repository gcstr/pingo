package main

import (
	"path/filepath"
	"testing"
	"time"
)

// Helper function to create a pointer to a float64
func float64Ptr(f float64) *float64 {
	return &f
}

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
		Min:       float64Ptr(10.5),
		Avg:       float64Ptr(12.3),
		Max:       float64Ptr(15.7),
		StdDev:    float64Ptr(2.1),
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
			Min:       float64Ptr(10.0 + float64(i)),
			Avg:       float64Ptr(12.0 + float64(i)),
			Max:       float64Ptr(15.0 + float64(i)),
			StdDev:    float64Ptr(2.0),
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
	if stats[0].Min != nil && stats[1].Min != nil && *stats[0].Min > *stats[1].Min {
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
			Min:       float64Ptr(10.0 + float64(i)),
			Avg:       float64Ptr(12.0 + float64(i)),
			Max:       float64Ptr(15.0 + float64(i)),
			StdDev:    float64Ptr(2.0),
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
	if stats[0].Min == nil || *stats[0].Min != 11.0 {
		t.Errorf("Expected first min 11.0, got %v", stats[0].Min)
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
		Min:       float64Ptr(10.0),
		Avg:       float64Ptr(12.0),
		Max:       float64Ptr(15.0),
		StdDev:    float64Ptr(2.0),
	}
	err = savePingStats(db, oldStats, 30)
	if err != nil {
		t.Fatalf("Failed to save old data: %v", err)
	}

	// Insert recent data
	recentStats := &PingStats{
		Timestamp: time.Now(),
		Min:       float64Ptr(11.0),
		Avg:       float64Ptr(13.0),
		Max:       float64Ptr(16.0),
		StdDev:    float64Ptr(2.5),
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

	if len(stats) > 0 && (stats[0].Min == nil || *stats[0].Min != 11.0) {
		t.Errorf("Expected recent data to remain, got min %v", stats[0].Min)
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
