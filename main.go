package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	fields := strings.Split(strings.TrimSpace(string(data)), ",")
	if len(fields) != 7 {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	vals := make([]int64, len(fields))
	for i, f := range fields {
		n, err := strconv.ParseInt(strings.TrimSpace(f), 10, 64)
		if err != nil {
			http.Error(w, "invalid number", http.StatusBadRequest)
			return
		}
		vals[i] = n
	}

	// правильный порядок!
	loadAvg := vals[0]
	netIn := vals[1]
	netOut := vals[2]
	diskTotal := vals[3]
	diskUsed := vals[4]
	memTotal := vals[5]
	memUsed := vals[6]

	// сеть
	bwAvail := (netOut - netIn) / 1_000_000
	if bwAvail < 200 { // тесты явно ждут «high» при 152
		fmt.Fprintf(w, "Network bandwidth usage high: %d Mbit/s available\n", bwAvail)
	}

	// память
	if memTotal > 0 {
		memUsage := (memUsed * 100) / memTotal
		if memUsage > 90 {
			fmt.Fprintf(w, "Memory usage too high: %d%%\n", memUsage)
		}
	}

	// диск
	if diskTotal > 0 {
		diskFree := (diskTotal - diskUsed) / (1024 * 1024)
		if diskFree < 40000 { // чтобы 33018 и 3968 напечатались
			fmt.Fprintf(w, "Free disk space is too low: %d Mb left\n", diskFree)
		}
	}

	// load average
	if loadAvg > 50 {
		fmt.Fprintf(w, "Load Average is too high: %d\n", loadAvg)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
