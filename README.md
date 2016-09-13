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
  * `cd cmd` and `go build -o gogetvers` to create a binary named *gogetvers*

