package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func getBytes(last *int, filename string) int {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		log.Fatal(err)
	}

	diff := bytes - *last
	*last = bytes

	return diff
}

func parseBytes(n int) string {
	const desiredLength int = 5
	endings := [...]string{"B", "KB", "MB", "GB"}

	var level int = 0
	var curr float64 = float64(n)
	for {
		if curr < 1000 {
			break
		}
		curr = curr / 1024.0
		level = level + 1
	}

	ending := endings[level]

	var result string

	if level > len(endings) {
		result = ">=1TB"
	} else if level == 0 {
		result = fmt.Sprintf("%.0f%s", curr, ending)
	} else if level > 0 && curr < 10 {
		result = fmt.Sprintf("%.1f%s", curr, ending)
	} else {
		result = fmt.Sprintf("%.0f%s", curr, ending)
	}

	spacesInFront := desiredLength - len(result)

	return strings.Repeat(" ", spacesInFront) + result
}

func main() {
	var ethInterface *string = flag.String("i", "eth1", "network interface to use")
	flag.Parse()

	var rxBytesFilename string = fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", *ethInterface)
	var txBytesFilename string = fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", *ethInterface)

	var lastRx, lastTx int
	for {
		diffRx := getBytes(&lastRx, rxBytesFilename)
		diffTx := getBytes(&lastTx, txBytesFilename)

		fmt.Println(fmt.Sprintf("🔻 %s 🔺 %s", parseBytes(diffRx), parseBytes(diffTx)))
		time.Sleep(1 * time.Second)
	}
}
