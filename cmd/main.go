package main

import (
	"errors"
	"fmt"
	gv "gogetvers"
	"os"
	"path/filepath"
)

var (
	goget *gv.GoGetVers
)

type options struct {
	path  string
	file  string
	dashg string
	dashm string
	dashn string
	dasht string
}

func main() {
	var err error
	opts := options{}
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()
	// Parse arguments.
	args := append([]string{}, os.Args[1:]...)
	if len(args) == 0 {
		dousage() // No args --> print usage.
		return
	}
	switch args[0] {
	case "-v":
		args = args[1:]
		doversion(false)
	case "--version":
		args = args[1:]
		doversion(true)
	case "-h", "--help":
		args = args[1:]
		dousage()
	case "checkout", "generate", "make", "print", "rebuild", "release", "tag":
		sub := args[0]
		args = args[1:]
		// Options parsing...
		for len(args) > 0 {
			curr := len(args)
			tempopts := []struct {
				flag   string
				target *string
			}{
				{"-f", &opts.file},
				{"-g", &opts.dashg},
				{"-m", &opts.dashm},
				{"-n", &opts.dashn},
				{"-t", &opts.dasht}}
			for _, opt := range tempopts {
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
				opts.path = args[0]
				args = args[1:]
			}
			if curr == len(args) {
				// Nothing done, so remove one to avoid infinite loop
				args = args[1:]
			}
		}
		// End options parsing.
		// Path wasn't provided
		if opts.path == "" {
			if sub == "checkout" {
				// Checkout uses GOPATH environment variable as default path.
				envpath := os.Getenv("GOPATH")
				if envpath == "" {
					fmt.Println("Error: GOPATH is not set.")
					exitCode = 1
					return
				}
			} else {
				// Everything else uses current directory.
				opts.path, err = os.Getwd()
				if err != nil {
					fmt.Println("Error:", err.Error())
					exitCode = 1
					return
				}
			}
		}
		// All commands require a path that exists
		if !gv.IsDir(opts.path) {
			fmt.Println(fmt.Sprintf("Error: PATH is not a directory: %v", opts.path))
			exitCode = 1
			return
		}
		// If file is not provided then default is "gogetvers.manifest" within PATH
		if opts.file == "" {
			opts.file = filepath.Join(opts.path, "gogetvers.manifest")
		}
		// The following commands require that FILE exists: checkout, generate, print, rebuild
		if sub == "checkout" || sub == "generate" || sub == "print" || sub == "rebuild" {
			if !gv.IsFile(opts.file) {
				fmt.Println(fmt.Sprintf("Error: FILE is not a file: %v", opts.file))
				exitCode = 1
				return
			}
		}
		// Create our GGV object.
		goget, err = gv.NewGoGetVers(opts.path, opts.file, os.Stdout)
		if err != nil {
			fmt.Printf("Error: %v\n", err.Error())
			exitCode = 1
			return
		}
		switch sub {
		case "checkout":
			err = docheckout()
		case "generate":
			err = dogenerate(opts.dashg, opts.dashn)
		case "make":
			err = domake()
		case "print":
			err = doprint()
		case "rebuild":
			err = dorebuild()
		case "release":
			err = dorelease(opts.dashg, opts.dashn, opts.dasht, opts.dashm)
		case "tag":
			err = dotag(opts.dasht)
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

func dorelease(gofile, packageName, tag, message string) error {
	if gofile == "" {
		gofile = filepath.Join(goget.Path, "generated_gogetvers.go")
	}
	if message == "" {
		message = tag + " by gogetvers"
	}
	return goget.Release(gofile, packageName, tag, message)
}

func dotag(tag string) error {
	return goget.Tag(tag)
}

func dogenerate(gofile, packageName string) error {
	if gofile == "" {
		gofile = filepath.Join(goget.Path, "generated_gogetvers.go")
	}
	return goget.Generate(gofile, packageName)
}

func doversion(long bool) {
	if long {
		fmt.Println(VersionInfo.GetVersionVerbose("gogetvers"))
	} else {
		fmt.Println(VersionInfo.GetVersion("gogetvers"))
	}
}

func dousage() {
	usage := `
gogetvers -v|--version
    Print version information.

gogetvers [-h|--help]
    Print help information.

For all proceeding commands:
    + If omitted then PATH defaults to current working directory,
      except for the checkout command where it defaults to the
      GOPATH environment variable.
    + If omitted then -f option defaults to gogetvers.manifest
      within PATH
    + If omitted then -g option defaults to generated_gogetvers.go
      within PATH

gogetvers checkout [-f MANIFEST] [PATH]
    Does the same as the 'rebuild' command with the following
    differences:
        + Uses GOPATH environment variable if PATH is omitted.
        + Will attempt to checkout the appropriate hash of
          a git dependency if it already exists on the file
          system.
    If any of the dependencies have local modifications then
    no work is performed.

gogetvers generate [-f MANIFEST] [-g GOFILE] [-n PACKAGENAME] [PATH]
    Create a go source file with version information at PATH
    using MANIFEST file.  GOFILE is the output filename or 
    generated_gogetvers.go if omitted.  By default PACKAGENAME will 
    be extracted from MANIFEST; use this option to specify another 
    name (i.e. for 'main').

gogetvers make [-f FILE] [PATH]
    Create manifest information for golang package at PATH; or
    in current directory if PATH is omitted. FILE can be used
    to specify the output location of the manifest information;
    default FILE is gogetvers.manifest in PATH.

gogetvers print [-f MANIFEST] | [PATH]
    Print a summary of the MANIFEST file in PATH.  PATH
    defaults to current directory; MANIFEST defaults to
    gogetvers.manifest.

gogetvers rebuild [-f MANIFEST] [PATH]
    Rebuild package structure described by MANIFEST at PATH;
    or in current directory if PATH is omitted.  If any of
    the dependencies described by MANIFEST already exist on
    the file system then no work is performed.

gogetvers release [-g GOFILE] [-n PACKAGENAME] [-m MESSAGE] -t TAG [PATH]
    Creates an annotated tag for a project.  The following
    commands are performed:
      + git tag -a TAG [-m MESSAGE]
      + git push origin TAG
      + gogetvers make PATH
      + gogetvers generate -g GOFILE -n PACKAGENAME PATH
      + git add . && git commit [-m MESSAGE]
    If omitted PATH will be the current directory.  Release
    requires that the project at PATH and all of its dependencies
    do not have local modifications.  This is a convenience
    command to make a release version of a package.

gogetvers tag -t TAG [PATH]
    Tag is similar to 'release' except the tag is not annotated and
    the check for local modifications is not performed.  This command
	is suitable for tagging development or feature branches.  The
    following commands are performed:
      + git tag -d TAG
      + git tag TAG
      + gogetvers make PATH

`
	fmt.Printf(usage)
}
