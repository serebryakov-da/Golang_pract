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

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-sig:
			break loop
		default:
			resp, err := http.Get("http://127.0.0.1/")
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			fields := strings.Split(strings.TrimSpace(string(body)), ",")
			if len(fields) != 7 {
				continue
			}

			vals := make([]int64, 7)
			for i, f := range fields {
				n, _ := strconv.ParseInt(strings.TrimSpace(f), 10, 64)
				vals[i] = n
			}

			currentLA := vals[0]
			memAvail := vals[1]
			memUsed := vals[2]
			diskAvail := vals[3]
			diskUsed := vals[4]
			netAvail := vals[5]
			netUsed := vals[6]

			// Load Average
			if currentLA > 30 {
				fmt.Printf("Load Average is too high: %d\n", currentLA)
			}

			// Memory
			if memAvail > 0 {
				usage := (memUsed * 100) / memAvail
				if usage >= 85 {
					fmt.Printf("Memory usage too high: %d%%\n", usage)
				}
			}

			// Disk
			if diskAvail > 0 {
				free := (diskAvail - diskUsed) / (1024 * 1024)
				if free < 10240 {
					fmt.Printf("Free disk space is too low: %d Mb left\n", free)
				}
			}

			// Network
			if netAvail > 0 {
				bw := (netAvail - netUsed) / 1_000_000
				if bw < 200 {
					fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", bw)
				}
			}

			time.Sleep(200 * time.Millisecond)
		}
	}
}
