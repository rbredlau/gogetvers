package main

import (
	"strings"
)

var VersionInfo = VersionInfoType{"edab0df1", []struct {
	Name    string
	Version string
}{{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"gogetvers", "edab0df1"},
	{"golang.org/x/exp", "3b75128c"},
	{"golang.org/x/net", "b6d7b139"}}}

type VersionInfoType struct {
	Version      string
	Dependencies []struct {
		Name    string
		Version string
	}
}

func (v VersionInfoType) GetVersion(binaryName string) string {
	return binaryName + " version " + v.Version
}

func (vt VersionInfoType) GetVersionVerbose(binaryName string) string {
	v := vt.GetVersion(binaryName)
	deps := []string{}
	for _, dep := range vt.Dependencies {
		deps = append(deps, dep.Name+" version "+dep.Version)
	}
	if len(deps) > 0 {
		v = v + "\n    " + strings.Join(deps, "\n    ")
	}
	return v
}
