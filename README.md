#gogitvers
Yet another package versioning tool for golang.

#What does it do?
gogitvers can embed git version information into your project and also
tracks version information of your project's dependencies.

#These tools already exist; why is this one so special?
It avoids all the nonsense with *vendor*, doesn't mangle import names
inside your existing project, doesn't otherwise copy dependencies
into your project, and uses the standard golang tools to do its job.

I consider this approach to be more *pure* in terms of golang's initial design concepts:
* Your project and its dependencies only exist under GOPATH and nowhere else.
* The existing golang tools (i.e. `go generate`) are used instead of introducing
makefiles or similar depedencies.

#How does it work?
gogetvers analyzes a golang package and its dependencies and generates a 
JSON formatted manifest file.  This manifest file can be used to embed
version information into your project and also to revert your project
and all its dependencies to prior states.

#Why is it two packages instead of one?
* gogetvers contains the code to do the heavy lifting.
* cmd contains the code to build a binary program.
  * `cd gogetvers/cmd` and `go build -o gogetvers` to create a binary named *gogetvers*

#### Documentation
[![](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/rbredlau/gogitvers)

#man gogetvers
```
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
```

#Examples

##gogetvers make
Makes the manifest file.
```
# Make manifest for myproject
$ cd $GOPATH/src/myproject
$ gogetvers make
# or
$ gogetvers make $GOPATH/src/myproject
```

##gogetvers print
Prints a summary of the manifest file.
```
# Print existing manifest in current directory.
$ cd $GOPATH/src/myproject
$ gogetvers print
# or
$ gogetvers print $GOPATH/src/myproject
# or
$ gogetvers print -f $GOPATH/src/myproject/gogetvers.manifest
```

##gogetvers rebuild
Clones the repositories from the manifest into a given path; none
of the repositories in the manifest can exist in the destination path.
```
$ mkdir foo
$ gogetvers rebuild -f $GOPATH/src/myproject/gogetvers.manifest foo
# or
$ mkdir bar
$ cd bar
$ gogetvers rebuild -f $GOPATH/src/myproject/gogetvers.manifest
```

##gogetvers checkout
The same as rebuild except the repositories from the manifest CAN exist on disk;
they will be checked out with the hash described in the manifest or cloned if
it doesn't exist on disk.  `checkout` can only be used if existing repositories
do not have local modifications.
```
$ cd $GOPATH/src/myproject
$ git checkout oldversion
$ gogetvers checkout
$ # Dependencies of myproject under $GOPATH will have be reverted to the
$ # hashes described in ./gogetvers.manifest
```

#@TODO
+ Implement `gogetvers tag`
+ `gogetvers make` should issue warnings if any package has local modifications.

