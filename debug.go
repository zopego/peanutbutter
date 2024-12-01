package panelbubble

import "fmt"

var (
	Debug     = false
	DebugChan = make(chan string, 100)
)

func DebugPrintf(format string, a ...any) {
	if Debug {
		DebugChan <- fmt.Sprintf(format, a...)
	}
}
