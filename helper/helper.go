package helper

import (
	"fmt"
)

func PrintUsage() {
	fmt.Println(`$ ./frappuccino --help
Coffee Shop Management System

Usage:
  frappuccino [--port <N>]
  frappuccino --help

Options:
  --help       Show this screen.
  --port N     Port number.`)
}
