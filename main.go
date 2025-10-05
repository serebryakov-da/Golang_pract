package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	serverURL     = "http://srv.msk01.gigacorp.local/_stats"
	defaultPeriod = 200 * time.Millisecond
	maxErrors     = 3
)

func main() {
	period := getInterval()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	errorCount := 0

loop:
	for {
		select {
		case <-sig:
			break loop
		case <-ticker.C:
			stats, err := fetchStats()
			if err != nil {
				errorCount++
				if errorCount >= maxErrors {
					fmt.Println("Unable to fetch server statistic")
					errorCount = 0
				}
				continue
			}
			// успешный ответ — сбрасываем счётчик ошибок
			errorCount = 0
			checkMetrics(stats)
		}
	}
}

func getInterval() time.Duration {
	if v := os.Getenv("CHECK_INTERVAL"); v != "" {
		if s, err := strconv.Atoi(v); err == nil && s > 0 {
			return time.Duration(s) * time.Second
		}
	}
	return defaultPeriod
}

func fetchStats() ([]int64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := strings.TrimSpace(string(data))
	parts := strings.Split(s, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format")
	}

	vals := make([]int64, 7)
	for i, p := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		vals[i] = n
	}
	return vals, nil
}

func checkMetrics(vals []int64) {

	currentLA := vals[0]
	memAvail := vals[1]
	memUsed := vals[2]
	diskAvail := vals[3]
	diskUsed := vals[4]
	netAvail := vals[5]
	netUsed := vals[6]

	if currentLA > 30 {
		fmt.Printf("Load Average is too high: %d\n", currentLA)
	}

	if memAvail > 0 {
		memUsage := (memUsed * 100) / memAvail // целочисленно, усечение
		if memUsage > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memUsage)
		}
	}

	if diskAvail > 0 {
		freeBytes := diskAvail - diskUsed
		freePercent := (freeBytes * 100) / diskAvail
		if freePercent < 10 {
			freeMb := freeBytes / (1024 * 1024) // перевод в МБ (целочисленно)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
		}
	}

	if netAvail > 0 {
		netUsagePercent := (netUsed * 100) / netAvail
		if netUsagePercent > 90 {
			availableBytes := netAvail - netUsed
			availableMbit := availableBytes / 1_000_000 // по тесту: /1000/1000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", availableMbit)
		}
	}
}
