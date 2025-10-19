package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type PingStats struct {
	Timestamp time.Time `json:"timestamp"`
	Min       float64   `json:"min"`
	Avg       float64   `json:"avg"`
	Max       float64   `json:"max"`
	StdDev    float64   `json:"stddev"`
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
		min REAL NOT NULL,
		avg REAL NOT NULL,
		max REAL NOT NULL,
		stddev REAL NOT NULL
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func savePingStats(db *sql.DB, stats *PingStats, retentionDays int) error {
	insertSQL := `INSERT INTO ping_stats (timestamp, min, avg, max, stddev) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, stats.Timestamp, stats.Min, stats.Avg, stats.Max, stats.StdDev)
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
	query := `SELECT timestamp, min, avg, max, stddev FROM ping_stats ORDER BY timestamp DESC LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []PingStats
	for rows.Next() {
		var s PingStats
		err := rows.Scan(&s.Timestamp, &s.Min, &s.Avg, &s.Max, &s.StdDev)
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

	query := `SELECT timestamp, min, avg, max, stddev FROM ping_stats
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
		err := rows.Scan(&s.Timestamp, &s.Min, &s.Avg, &s.Max, &s.StdDev)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}
