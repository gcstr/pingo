package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// validateTarget ensures the target is a valid hostname or IP address
// and doesn't contain shell metacharacters that could enable command injection
func validateTarget(target string) error {
	// Check for shell metacharacters
	if strings.ContainsAny(target, ";|&`$(){}[]<>\n\r") {
		return fmt.Errorf("invalid target: contains shell metacharacters")
	}

	// Try to parse as IP address
	if ip := net.ParseIP(target); ip != nil {
		return nil
	}

	// Validate as hostname (DNS name)
	// Hostnames can contain letters, digits, hyphens, and dots
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(target) {
		return fmt.Errorf("invalid target: not a valid hostname or IP address")
	}

	return nil
}

func runPing(target string, count int) (string, error) {
	// Validate target to prevent command injection
	if err := validateTarget(target); err != nil {
		return "", err
	}

	// Build ping command with platform-specific timeout flags
	args := []string{"-c", fmt.Sprint(count)}

	// Add timeout flag based on OS
	// macOS uses -t for TTL timeout
	// Linux uses -w for deadline (whole command timeout in seconds)
	if runtime.GOOS == "darwin" {
		args = append(args, "-t", fmt.Sprint(count))
	} else if runtime.GOOS == "linux" {
		args = append(args, "-w", fmt.Sprint(count))
	}
	// For other platforms, no timeout flag (Windows uses -w differently)

	args = append(args, target)

	cmd := exec.Command("ping", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func parsePingStats(output string) (*PingStats, error) {
	stats := &PingStats{
		Timestamp: time.Now(),
	}

	// Parse packet loss percentage
	// Example: "3 packets transmitted, 0 packets received, 100.0% packet loss"
	// or: "3 packets transmitted, 3 received, 0% packet loss"
	packetLossRe := regexp.MustCompile(`([0-9.]+)% packet loss`)
	packetLossMatches := packetLossRe.FindStringSubmatch(output)
	if len(packetLossMatches) >= 2 {
		packetLoss, _ := strconv.ParseFloat(packetLossMatches[1], 64)
		stats.PacketLoss = packetLoss
	} else {
		// If we can't parse packet loss at all (e.g., "cannot resolve host", timeout, etc.)
		// treat it as 100% packet loss
		stats.PacketLoss = 100.0
	}

	// Parse the summary line from both macOS and Linux formats:
	// macOS: round-trip min/avg/max/stddev = 9.713/11.739/13.595/1.849 ms
	// Linux: rtt min/avg/max/mdev = 8.106/8.247/8.387/0.106 ms
	// Note: stddev/mdev can be "nan" when only 1 packet received
	re := regexp.MustCompile(`(?:round-trip|rtt) min/avg/max/(?:stddev|mdev) = ([0-9.]+|nan)/([0-9.]+|nan)/([0-9.]+|nan)/([0-9.]+|nan) ms`)
	matches := re.FindStringSubmatch(output)

	if len(matches) != 5 {
		// If we can't parse ping stats (100% packet loss, timeout, or DNS failure)
		// Leave Min, Avg, Max, StdDev as nil (NULL in database)
		// This is valid - we have packet loss data but no latency data
		return stats, nil
	}

	// Parse values, treating "nan" as 0.0
	min, _ := strconv.ParseFloat(matches[1], 64)
	if math.IsNaN(min) {
		min = 0.0
	}
	avg, _ := strconv.ParseFloat(matches[2], 64)
	if math.IsNaN(avg) {
		avg = 0.0
	}
	max, _ := strconv.ParseFloat(matches[3], 64)
	if math.IsNaN(max) {
		max = 0.0
	}
	stddev, _ := strconv.ParseFloat(matches[4], 64)
	if math.IsNaN(stddev) {
		stddev = 0.0
	}

	stats.Min = &min
	stats.Avg = &avg
	stats.Max = &max
	stats.StdDev = &stddev

	return stats, nil
}

func runPingMonitor(db *sql.DB, target string, pingCount int, retentionDays int) {
	log.Printf("Starting continuous ping monitoring to %s with %d pings per round", target, pingCount)

	for {
		log.Printf("Running ping round...")
		output, cmdErr := runPing(target, pingCount)

		// Try to parse stats even if ping command failed
		// (output may still contain packet loss information)
		stats, err := parsePingStats(output)
		if err != nil {
			log.Printf("Failed to parse ping stats: %v (output: %s)", err, output)
			time.Sleep(5 * time.Second)
			continue
		}

		// Log the command error if there was one, but still save the stats
		if cmdErr != nil {
			log.Printf("Ping command error: %v (packet loss: %.1f%%)", cmdErr, stats.PacketLoss)
		}

		err = savePingStats(db, stats, retentionDays)
		if err != nil {
			log.Printf("Failed to save stats: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if stats.PacketLoss > 0 {
			if stats.Min != nil {
				log.Printf("Saved stats: min=%.3f avg=%.3f max=%.3f stddev=%.3f ms (packet loss: %.1f%%)",
					*stats.Min, *stats.Avg, *stats.Max, *stats.StdDev, stats.PacketLoss)
			} else {
				log.Printf("Saved stats: no data available (packet loss: %.1f%%)", stats.PacketLoss)
			}
		} else {
			log.Printf("Saved stats: min=%.3f avg=%.3f max=%.3f stddev=%.3f ms",
				*stats.Min, *stats.Avg, *stats.Max, *stats.StdDev)
		}

		// Wait before next ping round to avoid hammering the target
		time.Sleep(5 * time.Second)
	}
}
