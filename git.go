package gogetvers

import (
	fs "broadlux/fileSystem"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type Git struct {
	HomeDir   string
	Branch    string
	Hash      string
	OriginUrl string
	Describe  string
	Status    string
	//
	*pathsComposite
}

// Starts at path and works upwards looking for .git directory.
// Stops when it reaches stopDir and returns an error.
func findGitDir(path, stopDir string) (string, error) {
	if path == "" || !fs.IsDir(path) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", path))
	}
	if stopDir == "" || !fs.IsDir(stopDir) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", stopDir))
	}
	if stopDir == path {
		return "", errors.New(fmt.Sprintf("Search for git reached stopDir"))
	}
	try := filepath.Join(path, ".git")
	if fs.IsDir(try) {
		abs, err := filepath.Abs(try)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	return findGitDir(filepath.Dir(path), stopDir)
}

// Start at path and look upwards for a .git directory, stopping
// if stopDir is reached.  Return a git type from the found
// .git directory.
func newGitByFind(path, stopDir string) (*Git, error) {
	gitDir, err := findGitDir(path, stopDir)
	if err != nil {
		return nil, err
	}
	return NewGit(filepath.Dir(gitDir))
}

// Returns a new Git structure for the given path.
func NewGit(path string) (rv *Git, rverr error) {
	if !fs.IsDir(path) {
		return nil, errors.New(fmt.Sprintf("not a path @ %v", path))
	}
	if !fs.IsDir(filepath.Join(path, ".git")) {
		return nil, errors.New(fmt.Sprintf("path is not a git @ %v", path))
	}
	//
	rv = &Git{HomeDir: path}
	rv.pathsComposite = newPathsComposite(&rv.HomeDir)
	type tempIterator struct {
		command *command
		target  *string
	}
	commands := []tempIterator{
		tempIterator{newCommandGitBranch(), &rv.Branch},
		tempIterator{newCommandGitOrigin(), &rv.OriginUrl},
		tempIterator{newCommandGitHash(), &rv.Hash},
		tempIterator{newCommandGitStatus(), &rv.Status},
		tempIterator{newCommandGitDescribe(), &rv.Describe}}
	//
	for _, cmd := range commands {
		err := cmd.command.exec(path)
		if err == nil {
			*cmd.target = cmd.command.output
		}
	}
	if rverr != nil {
		return nil, rverr
	}
	//
	if rv.Branch != "" {
		pieces := strings.Split(rv.Branch, "\n")
		rv.Branch = ""
		for _, v := range pieces {
			v = strings.Trim(v, "\r\n ")
			if v[0] == '*' {
				rv.Branch = strings.TrimSpace(strings.Replace(strings.Replace(strings.Replace(v, "* ", "", -1), "(", "", -1), ")", "", -1))
				break
			}
		}
	}
	return rv, nil
}

// Returns git as a string.
func (g *Git) String() string {
	if g == nil {
		return ""
	}
	rv := g.HomeDir + "\n"
	rv = rv + "    origin> " + g.OriginUrl + "\n"
	rv = rv + "    branch> " + g.Branch + "\n"
	rv = rv + "    hash> " + g.Hash + "\n"
	rv = rv + "    describe> " + g.Describe + "\n"
	return rv
}
