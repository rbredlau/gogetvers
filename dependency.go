package gogetvers

import (
	"strings"
)

type Dependency interface {
	Dependency()
}

type DependencyComposite struct{}

func (d DependencyComposite) Dependency() {}

type BuiltinDependency struct {
	Name string
	// For dependency interface
	DependencyComposite
}

type GitDependency struct {
	Name string
	Git  *Git
	// For dependency interface
	DependencyComposite
}

type UntrackedDependency struct {
	Name string
	// For dependency interface
	DependencyComposite
}

func GetDependency(dependencyDir, rootDir string) (Dependency, error) {
	name := strings.Replace(dependencyDir, rootDir, "", 1)
	if !IsDir(dependencyDir) {
		// Must be a golang built in
		return &BuiltinDependency{Name: name, DependencyComposite: DependencyComposite{}}, nil
	}
	git, err := NewGitByFind(dependencyDir, rootDir)
	if err != nil {
		// Not a git repo so not trackable
		return &UntrackedDependency{Name: name, DependencyComposite: DependencyComposite{}}, nil
	}
	return &GitDependency{Name: name, Git: git, DependencyComposite: DependencyComposite{}}, nil
}
