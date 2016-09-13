#gogitvers#
Yet another package versioning tool for golang.

#What does it do?#
gogitvers can embed git version information into your project and also
tracks version information of your project's dependencies.

#These tools already exist; why is this one so special?#
It avoids all the nonsense with *vendor*, doesn't mangle import names
inside your existing project, doesn't otherwise copy dependencies
into your project, and uses the standard golang tools to do its job.

I consider this to be more *pure* in terms of golang's initial design concepts:
* Your project and its dependencies only exist under GOPATH and nowhere else.
* The existing golang tools (i.e. `go generate`) are used instead of introducing
makefiles or similar depedencies.
