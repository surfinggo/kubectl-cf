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
		fmt.Print(termenv.String(fmt.Sprintf("[DEBUG] "+format+"\n", a...)).Faint())
	}
}
