package main

import (
	"fmt"
	gv "gogetvers"
	"os"
)

var (
	cwd     string
	path    string
	file    string
	command string
)

func main() {
	// Get current path.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	// Path starts as current path.
	path = cwd
	// Parse arguments.
	args := append([]string{}, os.Args[1:]...)
	if len(args) == 0 {
		dousage() // No args --> print usage.
		os.Exit(0)
	}
	switch args[0] {
	case "-v", "--version":
		args = args[1:]
		doversion()
	case "-h", "--help":
		args = args[1:]
		dousage()
	case "const", "make", "rebuild":
		sub := args[0]
		args = args[1:]
		for len(args) > 0 {
			if args[0] == "-f" {
				if len(args) >= 2 {
					file = args[1]
					args = args[1:]
				} else {
					fmt.Println("Error: Missing value for -f")
					os.Exit(1)
				}
				args = args[1:]
			}
			if len(args) == 1 {
				path = args[0]
				args = args[1:]
			}
		}
		switch sub {
		case "const":
			doconst()
		case "make":
			domake()
		case "rebuild":
			dorebuild()
		}
	default:
		dousage() // Unknown args --> print usage.
	}
	os.Exit(0)
}

func domake() {
	if os.Args[1] != "make" {
		return
	}
	fmt.Println("domake") // TODO RM
	gv.Make("")
}

func dorebuild() {
	if os.Args[1] != "rebuild" {
		return
	}
	fmt.Println("dorebuild") // TODO RM
	gv.Rebuild("")
}

func doconst() {
	if os.Args[1] != "const" {
		return
	}
	fmt.Println("doconst") // TODO RM
	gv.Const("")
}

func doversion() {
	fmt.Printf("gogetvers version %v\n", "TODO: FILL ME IN")
}

func dousage() {
	usage := `
gogetvers -v|--version
    Print version information.

gogetvers [-h|--help]
    Print help information.

gogetvers make [-f FILE] [PATH]
    Create version information for golang package at PATH; or
    in current directory if PATH is omitted. FILE can be used
	to specify the output location of the version information.

gogetvers rebuild -f MANIFEST [PATH]
    Rebuild package structure described by MANIFEST at PATH;
    or in current directory if PATH is omitted.

gogetvers const [-f FILE] [PATH]
    Create a go source file with version information at PATH
    if provided or in current directory otherwise.  FILE can
    be used to specify the file name; if omitted file will
    be named generated_gogetvers.go.
`
	fmt.Printf(usage)
}
