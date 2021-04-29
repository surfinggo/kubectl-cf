package main

import (
	"fmt"
	"github.com/muesli/termenv"
	"os"
)

var (
	_debug = os.Getenv("DEBUG") != ""
)

func debug(format string, a ...interface{}) {
	if _debug {
		initialModel.meta = append(initialModel.meta,
			termenv.String(fmt.Sprintf("[DEBUG] "+format, a...)).Faint().String())
	}
}
