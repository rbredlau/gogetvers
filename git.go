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
	type tempIterator struct {
		command string
		args    []string
		target  *string
	}
	commands := []tempIterator{
		tempIterator{"git", []string{"branch"}, &rv.Branch},
		tempIterator{"git", []string{"config", "--get", "remote.origin.url"}, &rv.OriginUrl},
		tempIterator{"git", []string{"rev-parse", "HEAD"}, &rv.Hash},
		tempIterator{"git", []string{"status", "--porcelain"}, &rv.Status},
		tempIterator{"git", []string{"describe", "--tags", "--abbrev=8", "--always", "--long"}, &rv.Describe}}
	//
	for _, cmd := range commands {
		code, output, err := ExecProgram(path, cmd.command, cmd.args)
		if err != nil {
			rverr = err
		}
		if err == nil && code == 0 {
			output = strings.TrimSpace(output)
			*cmd.target = output
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

func (g *Git) StripDirPrefix(path string) {
	if g == nil {
		return
	}
	g.HomeDir = strings.TrimLeft(strings.Replace(g.HomeDir, path, "", -1), "\\/")
	g.ParentDir = strings.TrimLeft(strings.Replace(g.ParentDir, path, "", -1), "\\/")
}

func (g *Git) ToSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.ToSlash(g.HomeDir)
	g.ParentDir = filepath.ToSlash(g.ParentDir)
}

func (g *Git) FromSlash() {
	if g == nil {
		return
	}
	g.HomeDir = filepath.FromSlash(g.HomeDir)
	g.ParentDir = filepath.FromSlash(g.ParentDir)
}
