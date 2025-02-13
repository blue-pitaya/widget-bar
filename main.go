package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func cmdTrimmedOutput(cmd string) string {
	_cmd := exec.Command("bash", "-c", cmd)
	output, err := _cmd.CombinedOutput()
	if err != nil {
		return "?"
	}

	return strings.TrimSpace(string(output))
}

func printHeadsetBattery() string {
	battery := cmdTrimmedOutput("upower --dump | grep -A3 'headset' | grep 'percentage' | tr -d -c 0-9 | sed -e 's/$/%/'")

	result := "-"
	if battery != "" {
		result = battery
	}

	return " " + result
}

//  10%
func printSoundVolume() string {
	return cmdTrimmedOutput("amixer get Master | sed '$!d' | grep -E -o '[0-9]+%'")
}

func printRamUsage() string {
	return cmdTrimmedOutput("free -h | awk '/^Mem/ { print $3\"/\"$2 }' | sed s/i//g")
}

func prinTimeAndDate() string {
	return cmdTrimmedOutput("date '+%H:%M %d.%m.%Y'")
}

const version = "0.0.2"

func main() {
	var ethInterface *string = flag.String("i", "eth1", "network interface to use")

	versionFlag := flag.Bool("v", false, "Prints the version")
	versionFlagLong := flag.Bool("version", false, "Prints the version")

	flag.Parse()

	if *versionFlag || *versionFlagLong {
		fmt.Println(version)
		os.Exit(0)
	}

	var lastRx, lastTx int
	parts := [...]string{
		printNetworkTraffic(*ethInterface, &lastRx, &lastTx),
		printHeadsetBattery(),
		printSoundVolume(),
		printRamUsage(),
		prinTimeAndDate(),
	}
	output := strings.Join(parts[:], "  ")

	fmt.Printf("%s\n", output)
}
