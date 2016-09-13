package gogetvers

import (
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
	if path == "" || !IsDir(path) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", path))
	}
	if stopDir == "" || !IsDir(stopDir) {
		return "", errors.New(fmt.Sprintf("Not a path @ %v", stopDir))
	}
	if stopDir == path {
		return "", errors.New(fmt.Sprintf("Search for git reached stopDir"))
	}
	try := filepath.Join(path, ".git")
	if IsDir(try) {
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
	if !IsDir(path) {
		return nil, errors.New(fmt.Sprintf("not a path @ %v", path))
	}
	if !IsDir(filepath.Join(path, ".git")) {
		return nil, errors.New(fmt.Sprintf("path is not a git @ %v", path))
	}
	//
	rv = &Git{HomeDir: path}
	rv.setPathsComposite()
	type tempIterator struct {
		command *Command
		target  *string
	}
	commands := []tempIterator{
		tempIterator{NewCommandGitBranch(), &rv.Branch},
		tempIterator{NewCommandGitOrigin(), &rv.OriginUrl},
		tempIterator{NewCommandGitHash(), &rv.Hash},
		tempIterator{NewCommandGitStatus(), &rv.Status},
		tempIterator{NewCommandGitDescribe(), &rv.Describe}}
	//
	for _, cmd := range commands {
		err := cmd.command.Exec(path)
		if err == nil {
			*cmd.target = cmd.command.Output
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

// Sets the pathsComposite member.
func (g *Git) setPathsComposite() {
	if g != nil {
		g.pathsComposite = newPathsComposite(&g.HomeDir)
	}
}

// Clones the git.
func (g *Git) Clone(mkdirs bool) error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	parentDir := filepath.Dir(g.HomeDir)
	if !IsDir(parentDir) && mkdirs {
		err = Mkdir(parentDir, 0770)
		if err != nil {
			return err
		}
	}
	if !IsDir(parentDir) {
		err = errors.New(fmt.Sprintf("Not a dir @ %v", parentDir))
		return err
	}
	cmd := NewCommandGitClone("master", g.OriginUrl, filepath.Base(g.HomeDir))
	err = cmd.Exec(parentDir)
	if err != nil {
		return err
	}
	return nil
}

// Checksout the git to the proper hash.
func (g *Git) Checkout() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	if !IsDir(g.HomeDir) {
		err = errors.New(fmt.Sprintf("Not a dir @ %v", g.HomeDir))
		return err
	}
	cmd := NewCommandGitCheckout(g.Hash)
	err = cmd.Exec(g.HomeDir)
	if err != nil {
		return err
	}
	return nil
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
