package main

import (
	"encoding/json"
	"errors"
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

const (
	TimerStopped int = iota
	TimerRunning
	TimerPaused
)

type TimerState struct {
	Mode      int
	StartTime int64
}

func NewTimerState() *TimerState {
	return &TimerState{
		Mode:      TimerStopped,
		StartTime: 0,
	}
}

func saveTimerState(path string, state *TimerState) error {
	data, err := json.MarshalIndent(*state, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadTimerState(path string) (*TimerState, error) {
	state := NewTimerState()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return state, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}

	err = json.Unmarshal(data, &state)
	if err != nil {
		return state, err
	}

	return state, nil
}

func handleTimer(arg string) (string, error) {
	path := "/tmp/widget-bar_timer_state"
	state, err := loadTimerState(path)
	if err != nil {
		return "", err
	}

	switch arg {
	case "start":
		state.Mode = TimerRunning
		state.StartTime = time.Now().Unix()
		saveTimerState(path, state)
		return "", nil
	case "stop":
		state.Mode = TimerStopped
		saveTimerState(path, state)
		return "", nil
	case "get":
		switch state.Mode {
		case TimerStopped:
			return "", nil
		case TimerRunning:
			duration := time.Since(time.Unix(state.StartTime, 0))
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			seconds := int(duration.Seconds()) % 60

			res := fmt.Sprintf("\r%02d:%02d:%02d", hours, minutes, seconds)
			return res, nil
		case TimerPaused:
			return "", errors.New("eror")
		default:
			return "", errors.New("invalid timer state")
		}
	default:
		return "", errors.New("invalid timer arg")
	}

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
	var timerArgCommand *string = flag.String("timer-command", "get", "argument for timer")
	flag.Parse()

	timerOut, err := handleTimer(*timerArgCommand)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	versionFlag := flag.Bool("v", false, "Prints the version")
	versionFlagLong := flag.Bool("version", false, "Prints the version")

	if *versionFlag || *versionFlagLong {
		fmt.Println(version)
		os.Exit(0)
	}

	var lastRx, lastTx int
	parts := [...]string{
		timerOut,
		printNetworkTraffic(*ethInterface, &lastRx, &lastTx),
		printHeadsetBattery(),
		printSoundVolume(),
		printRamUsage(),
		prinTimeAndDate(),
	}
	var nonEmptyParts []string
	for _, part := range parts {
		if len(part) != 0 {
			nonEmptyParts = append(nonEmptyParts, part)
		}

	}
	output := strings.Join(nonEmptyParts, "  ")

	fmt.Printf("%s\n", output)
}
