package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func runPing(target string, count int) (string, error) {
	cmd := exec.Command("ping", "-c", fmt.Sprint(count), target)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func parsePingStats(output string) (*PingStats, error) {
	// Parse the summary line from both macOS and Linux formats:
	// macOS: round-trip min/avg/max/stddev = 9.713/11.739/13.595/1.849 ms
	// Linux: rtt min/avg/max/mdev = 8.106/8.247/8.387/0.106 ms
	re := regexp.MustCompile(`(?:round-trip|rtt) min/avg/max/(?:stddev|mdev) = ([0-9.]+)/([0-9.]+)/([0-9.]+)/([0-9.]+) ms`)
	matches := re.FindStringSubmatch(output)

	if len(matches) != 5 {
		return nil, fmt.Errorf("could not parse ping statistics from output")
	}

	min, _ := strconv.ParseFloat(matches[1], 64)
	avg, _ := strconv.ParseFloat(matches[2], 64)
	max, _ := strconv.ParseFloat(matches[3], 64)
	stddev, _ := strconv.ParseFloat(matches[4], 64)

	return &PingStats{
		Timestamp: time.Now(),
		Min:       min,
		Avg:       avg,
		Max:       max,
		StdDev:    stddev,
	}, nil
}

func runPingMonitor(db *sql.DB, target string, pingCount int, retentionDays int) {
	log.Printf("Starting continuous ping monitoring to %s with %d pings per round", target, pingCount)

	for {
		log.Printf("Running ping round...")
		output, err := runPing(target, pingCount)
		if err != nil {
			log.Printf("Ping failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		stats, err := parsePingStats(output)
		if err != nil {
			log.Printf("Failed to parse ping stats: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = savePingStats(db, stats, retentionDays)
		if err != nil {
			log.Printf("Failed to save stats: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Saved stats: min=%.3f avg=%.3f max=%.3f stddev=%.3f ms",
			stats.Min, stats.Avg, stats.Max, stats.StdDev)
	}
}
