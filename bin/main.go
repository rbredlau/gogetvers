package main

import (
	"fmt"
	gv "gogetvers"
	"os"
)

func main() {
	dousage(false)
	doversion()
	if len(os.Args) >= 2 {
		domake()
		dorebuild()
		doconst()
	}
	dousage(true)

	os.Exit(0)
}

func domake() {
	if os.Args[1] != "make" {
		return
	}
	fmt.Println("domake") // TODO RM
	gv.Make("")
	os.Exit(0)
}

func dorebuild() {
	if os.Args[1] != "rebuild" {
		return
	}
	fmt.Println("dorebuild") // TODO RM
	gv.Rebuild("")
	os.Exit(0)
}

func doconst() {
	if os.Args[1] != "const" {
		return
	}
	fmt.Println("doconst") // TODO RM
	gv.Const("")
	os.Exit(0)
}

func doversion() {
	for _, v := range os.Args {
		if v == "--version" {
			fmt.Printf("gogetvers version %v\n", "TODO: FILL ME IN")
			os.Exit(0)
		}
	}
}

func dousage(force bool) {
	usage := `
gogetvers --version
    Print version information.

gogetvers [-h|--help]
    Print help information.

gogetvers make [PATH]
    Create version information for golang package at PATH; or
    in current directory if PATH is omitted.

gogetvers rebuild -f MANIFEST [PATH]
    Rebuild package structure described by MANIFEST at PATH;
    or in current directory if PATH is omitted.

gogetvers const [-f FILE] [PATH]
    Create a go source file with version information at PATH
    if provided or in current directory otherwise.  FILE can
    be used to specify the file name; if omitted file will
    be named generated_gogetvers.go.
`
	for _, v := range os.Args {
		if v == "-h" || v == "--help" {
			fmt.Printf(usage)
			os.Exit(0)
		}
	}
	if len(os.Args) <= 1 || force {
		fmt.Printf(usage)
		os.Exit(0)
	}
}
