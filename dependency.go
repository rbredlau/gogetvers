package gogetvers

import (
	fs "broadlux/fileSystem"
	"strings"
)

type dependency interface {
	dependency()
}

type dependencyComposite struct{}

func (d dependencyComposite) dependency() {}

type builtinDependency struct {
	Name string
	// For dependency interface
	dependencyComposite
}

type gitDependency struct {
	Name string
	Git  *Git
	// For dependency interface
	dependencyComposite
}

type untrackedDependency struct {
	Name string
	// For dependency interface
	dependencyComposite
}

func getDependency(dependencyDir, rootDir string) (dependency, error) {
	name := strings.Replace(dependencyDir, rootDir, "", 1)
	if !fs.IsDir(dependencyDir) {
		// Must be a golang built in
		return &builtinDependency{Name: name, dependencyComposite: dependencyComposite{}}, nil
	}
	git, err := newGitByFind(dependencyDir, rootDir)
	if err != nil {
		// Not a git repo so not trackable
		return &untrackedDependency{Name: name, dependencyComposite: dependencyComposite{}}, nil
	}
	return &gitDependency{Name: name, Git: git, dependencyComposite: dependencyComposite{}}, nil
}
