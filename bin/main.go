package main

import (
	fs "broadlux/fileSystem"
	"errors"
	"fmt"
	gv "gogetvers"
	"os"
	"path/filepath"
)

var (
	goget *gv.GoGetVers
)

func main() {
	var cwd, path, file string
	var dashg, dashn string
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
	case "checkout", "const", "make", "print", "rebuild":
		sub := args[0]
		args = args[1:]
		parsedPath := false
		for len(args) > 0 {
			curr := len(args)
			opts := []struct {
				flag   string
				target *string
			}{
				{"-f", &file},
				{"-g", &dashg},
				{"-n", &dashn}}
			for _, opt := range opts {
				if len(args) > 0 && args[0] == opt.flag {
					if len(args) >= 2 {
						*opt.target = args[1]
						args = args[1:]
					} else {
						fmt.Printf("Error: Missing value for %v\n", opt.flag)
						exitCode = 1
						return
					}
					args = args[1:]
				}
			}
			//
			if len(args) == 1 {
				path = args[0]
				args = args[1:]
				parsedPath = true
			}
			if curr == len(args) {
				// Nothing done, so remove one to avoid infinite loop
				args = args[1:]
			}
		}
		if sub == "checkout" && !parsedPath {
			// TODO USE GOPATH
		}
		if !fs.IsDir(path) {
			fmt.Println(fmt.Sprintf("Error: PATH is not a directory: %v", path))
			exitCode = 1
			return
		}
		if sub != "make" && sub != "print" {
			if file != "" && !fs.IsFile(file) {
				fmt.Println(fmt.Sprintf("Error: FILE is not a file: %v", file))
				exitCode = 1
				return
			}
		} else if (sub == "make" || sub == "print") && file == "" {
			file = filepath.Join(path, "gogetvers.manifest")
		}
		goget, err = gv.NewGoGetVers(path, file, os.Stdout)
		if err != nil {
			fmt.Printf("Error: %v\n", err.Error())
			exitCode = 1
			return
		}
		switch sub {
		case "checkout":
			err = docheckout()
		case "const":
			err = doconst(dashg, dashn)
		case "make":
			err = domake()
		case "print":
			err = doprint()
		case "rebuild":
			err = dorebuild()
		default:
			err = errors.New("no sub command")
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
	return goget.Make()
}

func doprint() error {
	return goget.Print()
}

func docheckout() error {
	return goget.Checkout()
}

func dorebuild() error {
	return goget.Rebuild()
}

func doconst(gofile, packageName string) error {
	if gofile == "" {
		gofile = "generated_gogetvers.go"
	}
	return goget.Const(gofile, packageName)
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
    or in current directory if PATH is omitted.  If any of
    the dependencies described by MANIFEST already exists on
    the file system then no work is performed.

gogetvers checkout -f MANIFEST [PATH]
    Does the same as the 'rebuild' command with the following
    differences:
        + Uses GOPATH environment variable if PATH is omitted.
        + Will attempt to checkout the appropriate hash of
          a git dependency if it already exists on the file
          system.
    If any of the dependencies have local modifications then
    no work is performed.

gogetvers print [-f MANIFEST] | [PATH]
    Print a summary of the MANIFEST file in PATH.  PATH
    defaults to current directory; MANIFEST defaults to
    gogetvers.manifest.

gogetvers const -f MANIFEST [-g GOFILE] [-n PACKAGENAME] [PATH]
    Create a go source file with version information at PATH
    if provided or in current directory otherwise using MANIFEST
	file.  GOFILE is the output filename or generated_gogetvers.go
	if omitted.  By default PACKAGENAME will be extracted from
	MANIFEST; use this option to specify another name (i.e. for
	'main').
`
	fmt.Printf(usage)
}
