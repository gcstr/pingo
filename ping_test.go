package main

import (
	"testing"
)

func TestParsePingStatsMacOS(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: icmp_seq=0 ttl=118 time=10.234 ms
64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=11.456 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=118 time=9.789 ms

--- 8.8.8.8 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 9.789/10.493/11.456/0.693 ms`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Failed to parse macOS ping output: %v", err)
	}

	if stats.Min != 9.789 {
		t.Errorf("Expected min 9.789, got %f", stats.Min)
	}
	if stats.Avg != 10.493 {
		t.Errorf("Expected avg 10.493, got %f", stats.Avg)
	}
	if stats.Max != 11.456 {
		t.Errorf("Expected max 11.456, got %f", stats.Max)
	}
	if stats.StdDev != 0.693 {
		t.Errorf("Expected stddev 0.693, got %f", stats.StdDev)
	}
}

func TestParsePingStatsLinux(t *testing.T) {
	output := `PING google.com (172.217.168.206) 56(84) bytes of data.
64 bytes from ams16s32-in-f14.1e100.net (172.217.168.206): icmp_seq=1 ttl=118 time=8.39 ms
64 bytes from ams16s32-in-f14.1e100.net (172.217.168.206): icmp_seq=2 ttl=118 time=8.30 ms
64 bytes from ams16s32-in-f14.1e100.net (172.217.168.206): icmp_seq=3 ttl=118 time=8.11 ms
64 bytes from ams16s32-in-f14.1e100.net (172.217.168.206): icmp_seq=4 ttl=118 time=8.20 ms

--- google.com ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3004ms
rtt min/avg/max/mdev = 8.106/8.247/8.387/0.106 ms`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Failed to parse Linux ping output: %v", err)
	}

	if stats.Min != 8.106 {
		t.Errorf("Expected min 8.106, got %f", stats.Min)
	}
	if stats.Avg != 8.247 {
		t.Errorf("Expected avg 8.247, got %f", stats.Avg)
	}
	if stats.Max != 8.387 {
		t.Errorf("Expected max 8.387, got %f", stats.Max)
	}
	if stats.StdDev != 0.106 {
		t.Errorf("Expected mdev 0.106, got %f", stats.StdDev)
	}
}

func TestParsePingStatsInvalidOutput(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8): 56 data bytes
Request timeout for icmp_seq 0

--- 8.8.8.8 ping statistics ---
3 packets transmitted, 0 packets received, 100.0% packet loss`

	_, err := parsePingStats(output)
	if err == nil {
		t.Error("Expected error for invalid ping output, got nil")
	}
}

func TestParsePingStatsEmptyOutput(t *testing.T) {
	output := ""

	_, err := parsePingStats(output)
	if err == nil {
		t.Error("Expected error for empty ping output, got nil")
	}
}

func TestParsePingStatsWithWholeNumbers(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: icmp_seq=0 ttl=118 time=10 ms
64 bytes from 8.8.8.8: icmp_seq=1 ttl=118 time=11 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=118 time=9 ms

--- 8.8.8.8 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 9.0/10.0/11.0/1.0 ms`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Failed to parse ping output with whole numbers: %v", err)
	}

	if stats.Min != 9.0 {
		t.Errorf("Expected min 9.0, got %f", stats.Min)
	}
	if stats.Avg != 10.0 {
		t.Errorf("Expected avg 10.0, got %f", stats.Avg)
	}
	if stats.Max != 11.0 {
		t.Errorf("Expected max 11.0, got %f", stats.Max)
	}
	if stats.StdDev != 1.0 {
		t.Errorf("Expected stddev 1.0, got %f", stats.StdDev)
	}
}
