package main

import (
	"fmt"
	"os"
)

var (
	_debug = os.Getenv("DEBUG") != ""
)

func debug(format string, a ...interface{}) {
	if _debug {
		initialModel.meta = append(initialModel.meta, Subtle(fmt.Sprintf("[DEBUG] "+format, a...)))
	}
}
