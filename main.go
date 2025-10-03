package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL     = "http://srv.msk01.gigacorp.local/_stats"
	defaultPeriod = 30 * time.Second
	maxErrors     = 3
)

func main() {
	checkInterval := getInterval()
	errorCount := 0

	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(checkInterval)
			continue
		}

		errorCount = 0
		checkMetrics(stats)
		time.Sleep(checkInterval)
	}
}

func getInterval() time.Duration {
	if val := os.Getenv("CHECK_INTERVAL"); val != "" {
		if sec, err := strconv.Atoi(val); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return defaultPeriod
}

func fetchStats() ([]float64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}

	statsStr := strings.TrimSpace(scanner.Text())
	values := strings.Split(statsStr, ",")

	if len(values) != 7 {
		return nil, fmt.Errorf("invalid data format")
	}

	var stats []float64
	for _, val := range values {
		num, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return nil, err
		}
		stats = append(stats, num)
	}

	return stats, nil
}

func checkMetrics(stats []float64) {
	// 0 - Load Average
	if stats[0] > 30 {
		fmt.Printf("Load Average is too high: %.0f\n", stats[0])
	}

	// 1 - Total RAM, 2 - Used RAM (>80%)
	if stats[1] > 0 {
		memoryUsagePercent := int((stats[2] / stats[1]) * 100) // усечение без округления
		if memoryUsagePercent > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memoryUsagePercent)
		}
	}

	// 3 - Total Disk, 4 - Used Disk (<10% free)
	if stats[3] > 0 {
		freeDiskBytes := stats[3] - stats[4]
		freeDiskPercent := (freeDiskBytes / stats[3]) * 100
		if freeDiskPercent < 10 {
			// перевод в мегабайты через 10^6
			freeDiskMB := freeDiskBytes / 1_000_000
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDiskMB)
		}
	}

	// 5 - Total Network, 6 - Used Network (>90%)
	if stats[5] > 0 {
		networkUsagePercent := (stats[6] / stats[5]) * 100
		if networkUsagePercent > 90 {
			availableBandwidthBytes := stats[5] - stats[6]
			// байты -> мегабайты через 10^6
			availableBandwidthMbps := availableBandwidthBytes / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", availableBandwidthMbps)
		}
	}
}
