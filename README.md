##gogitvers
Yet another package versioning tool for golang.

##What does it do?
gogitvers can embed git version information into your project and also
tracks version information of your project's dependencies.

##These tools already exist; why is this one so special?
It avoids all the nonsense with *vendor*, doesn't mangle import names
inside your existing project, doesn't otherwise copy dependencies
into your project, and uses the standard golang tools to do its job.

I consider this approach to be more *pure* in terms of golang's initial design concepts:
* Your project and its dependencies only exist under GOPATH and nowhere else.
* The existing golang tools (i.e. `go generate`) are used instead of introducing
makefiles or similar depedencies.

##How does it work?
gogetvers analyzes a golang package and its dependencies and generates a 
JSON formatted manifest file.  This manifest file can be used to embed
version information into your project and also to revert your project
and all its dependencies to prior states.

##Why is it two packages instead of one?
* gogetvers contains the code to do the heavy lifting.
* cmd contains the code to build a binary program.
  * `cd gogetvers/cmd` and `go build -o gogetvers` to create a binary named *gogetvers*

## Documentation
[![](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/rbredlau/gogitvers)

##man gogetvers
```
gogetvers -v|--version
    Print version information.

gogetvers [-h|--help]
    Print help information.

gogetvers checkout -f MANIFEST [PATH]
    Does the same as the 'rebuild' command with the following
    differences:
        + Uses GOPATH environment variable if PATH is omitted.
        + Will attempt to checkout the appropriate hash of
          a git dependency if it already exists on the file
          system.
    If any of the dependencies have local modifications then
    no work is performed.

gogetvers generate -f MANIFEST [-g GOFILE] [-n PACKAGENAME] [PATH]
    Create a go source file with version information at PATH
    if provided or in current directory otherwise using MANIFEST
    file.  GOFILE is the output filename or generated_gogetvers.go
    if omitted.  By default PACKAGENAME will be extracted from
    MANIFEST; use this option to specify another name (i.e. for
    'main').

gogetvers make [-f FILE] [PATH]
    Create manifest information for golang package at PATH; or
    in current directory if PATH is omitted. FILE can be used
    to specify the output location of the manifest information;
    default FILE is gogetvers.manifest in PATH.

gogetvers print [-f MANIFEST] | [PATH]
    Print a summary of the MANIFEST file in PATH.  PATH
    defaults to current directory; MANIFEST defaults to
    gogetvers.manifest.

gogetvers rebuild -f MANIFEST [PATH]
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
    the check for local modifications is not performed.  'tag' is
    suitable for tagging development or feature branches.  The
    following commands are performed:
      + git tag -d TAG
      + git tag TAG
      + gogetvers make PATH
```

##Examples

###gogetvers make
Makes the manifest file.
```
# Make manifest for myproject
$ cd $GOPATH/src/myproject
$ gogetvers make
```
*or*
```
$ gogetvers make $GOPATH/src/myproject
```
*or*
```
$ cd $GOPATH/src/myproject
$ gogetvers make -f ~/current.manifest 
```
*or*
```
$ gogetvers make -f ~/current.manifest $GOPATH/src/myproject
```

###gogetvers print
Prints a summary of the manifest file.
```
# Print existing manifest in current directory.
$ cd $GOPATH/src/myproject
$ gogetvers print
```
*or*
```
$ gogetvers print $GOPATH/src/myproject
```
*or*
```
$ gogetvers print -f $GOPATH/src/myproject/gogetvers.manifest
```

###gogetvers rebuild
Clones the repositories from the manifest into a given path; none
of the repositories in the manifest can exist in the destination path.
```
$ mkdir foo
$ gogetvers rebuild -f $GOPATH/src/myproject/gogetvers.manifest foo
```
*or*
```
$ mkdir bar
$ cd bar
$ gogetvers rebuild -f $GOPATH/src/myproject/gogetvers.manifest
```

###gogetvers checkout
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

###gogetvers generate
This generates a golang source file with a `type VersionInfoType struct` and a 
single global variable named `VersionInfo` that contains the version information
contained in a manifest file.

`VersionInfo` has two public methods:
+ `GetVersion()` returns the version information for the primary package.
+ `GetVersionVerbose()` returns version information for the package and all dependencies.
```
$ gogetvers generate -f gogetvers.manifest 
```
*or*
```
$ gogetvers generate $gopath/src/myproject
```
*or*
```
$ gogetvers generate -f $GOPATH/src/myproject/gogetvers.manifest $gopath/src/myproject
```
By default `generate` uses the package name found in manifest; if you need to provide
a different name (such as `main`), then use the `-n` option:
```
$ gogetvers generate -f $GOPATH/src/myproject/gogetvers.manifest -n main $gopath/src/myproject
```

##This looks great but there's a HUGE problem...
gogetvers doesn't make a *deep copy* of dependencies.  If the git repositories
move or disappear then gogetvers can't `rebuild` or `checkout` old versions.  (*You
can always edit the manifest by hand to point at new locations if necessary though.*)

I think this is OK.

The most common use case for checking out old code is to duplicate a bug to make
a fix.  Most often this happens with recent versions - therefore access to the
dependencies in a manifest file will probably still be available.

If a dependency for your project disappears entirely from the internet then
there's a good chance it is no longer maintained and you should be looking for
a suitable replacement anyways.  If this happens and a bug is reported for
an old version you will most likely want to update the bug reporter to the newest
version anyways.

And finally - if you must always be able to rebuild the structure described in a
gogetvers manifest - you can always use `git clone --mirror` to clone a repository
to a location that's always available to you, optionally performing backups on
that location.

If you absolutely must immortalize and forever make available everything your
project was built with - or if you disagree with the reasoning given - then
gogetvers is not for you.

##Known bugs
gogetvers considers a dependency *trackable* if it has a .git directory in its root 
directory or in any of its parent directories.  If the .git directory is 
in a parent directory that excludes the dependency via .gitignore then gogetvers 
considers the dependency *tracked* even though it is ignored by source code control.

##@TODO
+ Get package name automatically with: `go list -f {{Name}}`

