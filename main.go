package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL     = "http://srv.msk01.gigacorp.local/_stats"
	checkInterval = 30 * time.Second
	maxErrors     = 3
)

func main() {
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
	// Load Average
	if stats[0] > 30 {
		fmt.Printf("Load Average is too high: %.0f\n", math.Round(stats[0]))
	}
	
	// Memory usage
	if stats[1] > 0 {
		memoryUsagePercent := (stats[2] / stats[1]) * 100
		if memoryUsagePercent > 80 {
			fmt.Printf("Memory usage too high: %.0f%%\n", math.Round(memoryUsagePercent))
		}
	}
	
	// Disk space
	if stats[3] > 0 {
		freeDiskBytes := stats[3] - stats[4]
		freeDiskPercent := (freeDiskBytes / stats[3]) * 100
		if freeDiskPercent < 10 {
			freeDiskMB := freeDiskBytes / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", math.Floor(freeDiskMB))
		}
	}
	
	// Network bandwidth
	if stats[5] > 0 {
		networkUsagePercent := (stats[6] / stats[5]) * 100
		if networkUsagePercent > 90 {
			availableBandwidthBits := (stats[5] - stats[6]) * 8
			availableBandwidthMbps := availableBandwidthBits / (1000 * 1000)
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", math.Round(availableBandwidthMbps))
		}
	}
}