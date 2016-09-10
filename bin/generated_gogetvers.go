package main

var VersionInfo = VersionInfoType{"3962a236", []struct {
	Name    string
	Version string
}{{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"broadlux", "1.0.1-2-g6c7bc7ce"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"github.com/stretchr/testify", "v1.1.3-16-g6cb3b85e"},
	{"gogetvers", "3962a236"},
	{"golang.org/x/exp", "3b75128c"},
	{"golang.org/x/net", "b6d7b139"}}}

type VersionInfoType struct {
	Version      string
	Dependencies []struct {
		Name    string
		Version string
	}
}
