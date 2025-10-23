package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"

	_ "modernc.org/sqlite"
)

type PingStats struct {
	Timestamp  time.Time  `json:"timestamp"`
	Min        *float64   `json:"min"`                 // Nullable - NULL when no data available
	Avg        *float64   `json:"avg"`                 // Nullable - NULL when no data available
	Max        *float64   `json:"max"`                 // Nullable - NULL when no data available
	StdDev     *float64   `json:"stddev"`              // Nullable - NULL when no data available
	PacketLoss float64    `json:"packet_loss"`         // Percentage 0-100
}

func initDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ping_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		min REAL,
		avg REAL,
		max REAL,
		stddev REAL,
		packet_loss REAL DEFAULT 0
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	// Add packet_loss column to existing tables if it doesn't exist
	alterTableSQL := `ALTER TABLE ping_stats ADD COLUMN packet_loss REAL DEFAULT 0`
	_, _ = db.Exec(alterTableSQL) // Ignore error if column already exists

	// Migrate existing tables: SQLite doesn't support ALTER COLUMN to drop NOT NULL
	// We need to recreate the table if it has NOT NULL constraints
	// Check if we need to migrate by looking at the table schema
	var sql string
	err = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='ping_stats'").Scan(&sql)
	if err == nil {
		// Check if the old schema has "NOT NULL" for min column
		if regexp.MustCompile(`min\s+REAL\s+NOT NULL`).MatchString(sql) {
			log.Printf("Migrating database schema to allow NULL values for ping failures...")

			// Create new table with correct schema
			_, err = db.Exec(`
				CREATE TABLE ping_stats_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					timestamp DATETIME NOT NULL,
					min REAL,
					avg REAL,
					max REAL,
					stddev REAL,
					packet_loss REAL DEFAULT 0
				)
			`)
			if err != nil {
				return nil, fmt.Errorf("failed to create new table: %v", err)
			}

			// Copy data from old table
			_, err = db.Exec(`
				INSERT INTO ping_stats_new (id, timestamp, min, avg, max, stddev, packet_loss)
				SELECT id, timestamp, min, avg, max, stddev, COALESCE(packet_loss, 0)
				FROM ping_stats
			`)
			if err != nil {
				return nil, fmt.Errorf("failed to copy data: %v", err)
			}

			// Drop old table
			_, err = db.Exec(`DROP TABLE ping_stats`)
			if err != nil {
				return nil, fmt.Errorf("failed to drop old table: %v", err)
			}

			// Rename new table
			_, err = db.Exec(`ALTER TABLE ping_stats_new RENAME TO ping_stats`)
			if err != nil {
				return nil, fmt.Errorf("failed to rename table: %v", err)
			}

			log.Printf("Database schema migration completed successfully")
		}
	}

	// Create index on timestamp for better query performance
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ping_stats_timestamp ON ping_stats(timestamp)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create timestamp index: %v", err)
	}

	return db, nil
}

func savePingStats(db *sql.DB, stats *PingStats, retentionDays int) error {
	insertSQL := `INSERT INTO ping_stats (timestamp, min, avg, max, stddev, packet_loss) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, stats.Timestamp, stats.Min, stats.Avg, stats.Max, stats.StdDev, stats.PacketLoss)
	if err != nil {
		return err
	}

	// Delete records older than retention period
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	deleteSQL := `DELETE FROM ping_stats WHERE timestamp < ?`
	_, err = db.Exec(deleteSQL, cutoffTime)
	return err
}

func getRecentStats(db *sql.DB, limit int) ([]PingStats, error) {
	query := `SELECT timestamp, min, avg, max, stddev, COALESCE(packet_loss, 0) FROM ping_stats ORDER BY timestamp DESC LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PingStats
	for rows.Next() {
		var s PingStats
		// Scan into pointers - NULL values will result in nil pointers
		err := rows.Scan(&s.Timestamp, &s.Min, &s.Avg, &s.Max, &s.StdDev, &s.PacketLoss)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	// Reverse to get chronological order
	for i := 0; i < len(stats)/2; i++ {
		j := len(stats) - 1 - i
		stats[i], stats[j] = stats[j], stats[i]
	}

	return stats, nil
}

func getStatsByDateRange(db *sql.DB, startDate, endDate string) ([]PingStats, error) {
	// Parse the input dates and convert to the format SQLite uses
	startTime, err := time.Parse("2006-01-02T15:04:05", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %v", err)
	}
	endTime, err := time.Parse("2006-01-02T15:04:05", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %v", err)
	}

	query := `SELECT timestamp, min, avg, max, stddev, COALESCE(packet_loss, 0) FROM ping_stats
	          WHERE timestamp >= ? AND timestamp <= ?
	          ORDER BY timestamp ASC`
	rows, err := db.Query(query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PingStats
	for rows.Next() {
		var s PingStats
		// Scan into pointers - NULL values will result in nil pointers
		err := rows.Scan(&s.Timestamp, &s.Min, &s.Avg, &s.Max, &s.StdDev, &s.PacketLoss)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func getStatsSince(db *sql.DB, since string) ([]PingStats, error) {
	// Parse the timestamp
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return nil, fmt.Errorf("invalid since timestamp format: %v", err)
	}

	query := `SELECT timestamp, min, avg, max, stddev, COALESCE(packet_loss, 0) FROM ping_stats
	          WHERE timestamp > ?
	          ORDER BY timestamp ASC`
	rows, err := db.Query(query, sinceTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PingStats
	for rows.Next() {
		var s PingStats
		// Scan into pointers - NULL values will result in nil pointers
		err := rows.Scan(&s.Timestamp, &s.Min, &s.Avg, &s.Max, &s.StdDev, &s.PacketLoss)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}
