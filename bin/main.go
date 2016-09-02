package main

import (
	"fmt"
	gv "gogetvers"
	"io"
	"os"
	"path/filepath"
)

var (
	cwd     string
	path    string
	file    string
	command string
	writer  io.Writer
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()
	// Get current path.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err.Error())
		exitCode = 1
		return
	}
	// Path starts as current path.
	path = cwd
	// Parse arguments.
	args := append([]string{}, os.Args[1:]...)
	if len(args) == 0 {
		dousage() // No args --> print usage.
		return
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
					exitCode = 1
					return
				}
				args = args[1:]
			}
			if len(args) == 1 {
				path = args[0]
				args = args[1:]
			}
		}
		if !gv.IsDir(path) {
			fmt.Println(fmt.Sprintf("Error: PATH is not a directory: %v", path))
			exitCode = 1
			return
		}
		if sub != "make" {
			if file != "" && !gv.IsFile(file) {
				fmt.Println(fmt.Sprintf("Error: FILE is not a file: %v", file))
				exitCode = 1
				return
			}
		} else if sub == "make" && file == "" {
			file = filepath.Join(path, "gogetvers.manifest")
		}
		writer = os.Stdout
		switch sub {
		case "const":
			err = doconst()
		case "make":
			err = domake()
		case "rebuild":
			err = dorebuild()
		}
		if err != nil {
			fmt.Println("Error:", err.Error())
			exitCode = 1
			return
		}
	default:
		dousage() // Unknown args --> print usage.
	}
}

func domake() error {
	return gv.Make(path, file, writer)
}

func dorebuild() error {
	return gv.Rebuild(path, file)
}

func doconst() error {
	return gv.Const(path, file)
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
    Create manifest information for golang package at PATH; or
    in current directory if PATH is omitted. FILE can be used
    to specify the output location of the manifest information;
    default FILE is gogetvers.manifest in PATH.

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
