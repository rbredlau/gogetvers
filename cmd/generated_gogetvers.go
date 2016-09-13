package main

import (
	"strings"
)

var VersionInfo = VersionInfoType{"a3f9df98", []struct {
	Name    string
	Version string
}{{"gogetvers", "a3f9df98"}}}

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
