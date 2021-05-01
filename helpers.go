package main

import (
	"fmt"
	"os"
)

var (
	debug = os.Getenv("DEBUG") != ""
)

// addDebugMessage adds a debug message to meta, which will be displayed on top of the output
func addDebugMessage(format string, a ...interface{}) {
	if debug {
		initialModel.meta = append(initialModel.meta, Subtle(fmt.Sprintf("[DEBUG] "+format, a...)))
	}
}
