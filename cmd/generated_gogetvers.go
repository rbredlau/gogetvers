package main

import (
	"strings"
)

// Global variable containing version information from
// gogetvers.
var VersionInfo = VersionInfoType{"dev-0-g40dd4a14", []struct {
	Name    string
	Version string
}{{"gogetvers", "dev-0-g40dd4a14"}}}

// Contains version information for package and its dependencies.
type VersionInfoType struct {
	Version      string
	Dependencies []struct {
		Name    string
		Version string
	}
}

// Returns the version for the package.
func (vt VersionInfoType) GetVersion(binaryName string) string {
	return binaryName + " version " + vt.Version
}

// Returns the version for the package and all of its dependencies.
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
