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

	if stats.Min == nil || *stats.Min != 9.789 {
		t.Errorf("Expected min 9.789, got %v", stats.Min)
	}
	if stats.Avg == nil || *stats.Avg != 10.493 {
		t.Errorf("Expected avg 10.493, got %v", stats.Avg)
	}
	if stats.Max == nil || *stats.Max != 11.456 {
		t.Errorf("Expected max 11.456, got %v", stats.Max)
	}
	if stats.StdDev == nil || *stats.StdDev != 0.693 {
		t.Errorf("Expected stddev 0.693, got %v", stats.StdDev)
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

	if stats.Min == nil || *stats.Min != 8.106 {
		t.Errorf("Expected min 8.106, got %v", stats.Min)
	}
	if stats.Avg == nil || *stats.Avg != 8.247 {
		t.Errorf("Expected avg 8.247, got %v", stats.Avg)
	}
	if stats.Max == nil || *stats.Max != 8.387 {
		t.Errorf("Expected max 8.387, got %v", stats.Max)
	}
	if stats.StdDev == nil || *stats.StdDev != 0.106 {
		t.Errorf("Expected mdev 0.106, got %v", stats.StdDev)
	}
}

func TestParsePingStatsInvalidOutput(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8): 56 data bytes
Request timeout for icmp_seq 0

--- 8.8.8.8 ping statistics ---
3 packets transmitted, 0 packets received, 100.0% packet loss`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Expected no error for 100%% packet loss, got: %v", err)
	}
	if stats.PacketLoss != 100.0 {
		t.Errorf("Expected packet loss 100.0, got %f", stats.PacketLoss)
	}
	if stats.Min != nil || stats.Avg != nil || stats.Max != nil || stats.StdDev != nil {
		t.Errorf("Expected nil stats for 100%% packet loss, got min=%v avg=%v max=%v stddev=%v",
			stats.Min, stats.Avg, stats.Max, stats.StdDev)
	}
}

func TestParsePingStatsEmptyOutput(t *testing.T) {
	output := ""

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Expected no error for empty ping output (should return 100%% packet loss), got: %v", err)
	}
	if stats.PacketLoss != 100.0 {
		t.Errorf("Expected packet loss 100.0 for empty output, got %f", stats.PacketLoss)
	}
	if stats.Min != nil || stats.Avg != nil || stats.Max != nil || stats.StdDev != nil {
		t.Errorf("Expected nil stats for empty output, got min=%v avg=%v max=%v stddev=%v",
			stats.Min, stats.Avg, stats.Max, stats.StdDev)
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

	if stats.Min == nil || *stats.Min != 9.0 {
		t.Errorf("Expected min 9.0, got %v", stats.Min)
	}
	if stats.Avg == nil || *stats.Avg != 10.0 {
		t.Errorf("Expected avg 10.0, got %v", stats.Avg)
	}
	if stats.Max == nil || *stats.Max != 11.0 {
		t.Errorf("Expected max 11.0, got %v", stats.Max)
	}
	if stats.StdDev == nil || *stats.StdDev != 1.0 {
		t.Errorf("Expected stddev 1.0, got %v", stats.StdDev)
	}
}

func TestParsePingStatsCannotResolveHost(t *testing.T) {
	output := `ping: cannot resolve google.com: Unknown host`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Expected no error for 'cannot resolve host' (should return 100%% packet loss), got: %v", err)
	}
	if stats.PacketLoss != 100.0 {
		t.Errorf("Expected packet loss 100.0 for DNS failure, got %f", stats.PacketLoss)
	}
	if stats.Min != nil || stats.Avg != nil || stats.Max != nil || stats.StdDev != nil {
		t.Errorf("Expected nil stats for DNS failure, got min=%v avg=%v max=%v stddev=%v",
			stats.Min, stats.Avg, stats.Max, stats.StdDev)
	}
}

func TestParsePingStatsWithNaN(t *testing.T) {
	output := `PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: icmp_seq=0 ttl=118 time=105.910 ms
ping: sendto: No route to host
Request timeout for icmp_seq 1

--- 8.8.8.8 ping statistics ---
5 packets transmitted, 1 packets received, 80.0% packet loss
round-trip min/avg/max/stddev = 105.910/105.910/105.910/nan ms`

	stats, err := parsePingStats(output)
	if err != nil {
		t.Fatalf("Failed to parse ping output with nan: %v", err)
	}

	if stats.PacketLoss != 80.0 {
		t.Errorf("Expected packet loss 80.0, got %f", stats.PacketLoss)
	}
	if stats.Min == nil || *stats.Min != 105.910 {
		t.Errorf("Expected min 105.910, got %v", stats.Min)
	}
	if stats.Avg == nil || *stats.Avg != 105.910 {
		t.Errorf("Expected avg 105.910, got %v", stats.Avg)
	}
	if stats.Max == nil || *stats.Max != 105.910 {
		t.Errorf("Expected max 105.910, got %v", stats.Max)
	}
	// nan should be converted to 0.0
	if stats.StdDev == nil {
		t.Errorf("Expected stddev 0.0 (from nan), got nil")
	} else if *stats.StdDev != 0.0 {
		t.Errorf("Expected stddev 0.0 (from nan), got %f", *stats.StdDev)
	}
}
