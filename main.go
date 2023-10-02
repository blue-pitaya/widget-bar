package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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

// this should work like: numfmt --to=iec $n
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

// 🔻 1.1KB 🔺  439B
func printNetworkTraffic(ethInterface string, lastRx *int, lastTx *int) string {
	var rxBytesFilename string = fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ethInterface)
	var txBytesFilename string = fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ethInterface)

	diffRx := getBytes(lastRx, rxBytesFilename)
	diffTx := getBytes(lastTx, txBytesFilename)

	return fmt.Sprintf("|🔻 %s 🔺 %s|", parseBytes(diffRx), parseBytes(diffTx))
}

//  10%
func printSoundVolume() string {
	cmd := exec.Command("bash", "-c", "amixer get Master | sed '$!d' | grep -E -o '[0-9]+%'")
	output, err := cmd.CombinedOutput()

	var result string

	if err != nil {
		result = "?"
	}

	result = strings.TrimSpace(string(output))

	return fmt.Sprintf(" %s", result)
}

func printRamUsage() string {
	cmd := exec.Command("bash", "-c", "free -h | awk '/^Mem/ { print $3\"/\"$2 }' | sed s/i//g")
	output, err := cmd.CombinedOutput()

	var result string

	if err != nil {
		result = "?"
	}

	result = strings.TrimSpace(string(output))

	return result
}

func main() {
	var ethInterface *string = flag.String("i", "eth1", "network interface to use")
	flag.Parse()

	var lastRx, lastTx int

	for {
		parts := [...]string{
			printNetworkTraffic(*ethInterface, &lastRx, &lastTx),
			printSoundVolume(),
			printRamUsage(),
		}
		output := strings.Join(parts[:], "  ")

		fmt.Println(output)

		time.Sleep(1 * time.Second)
	}
}
