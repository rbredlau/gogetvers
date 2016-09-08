package gogetvers

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type Git struct {
	HomeDir   string
	ParentDir string
	Branch    string
	Hash      string
	OriginUrl string
	Describe  string
	Status    string
	//
	*pathsComposite
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
	rv = &Git{HomeDir: path, ParentDir: filepath.Dir(path)}
	rv.pathsComposite = newPathsComposite(&rv.HomeDir, &rv.ParentDir)
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
