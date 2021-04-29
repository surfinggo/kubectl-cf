package main

import "github.com/muesli/termenv"

var (
	ColorProfile = termenv.ColorProfile()
	Warning      = ColorProfile.Color("1")
	Info         = ColorProfile.Color("28")
)
