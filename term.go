package main

import "github.com/muesli/termenv"

var (
	ColorProfile = termenv.ColorProfile()
	Warning      = makeFgStyle("1")
	Info         = makeFgStyle("28")
	Subtle       = makeFgStyle("241")
)

// Return a function that will colorize the foreground of a given string.
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(ColorProfile.Color(color)).Styled
}
