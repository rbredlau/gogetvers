package gogetvers

const version_template = `
package $PACKAGE_NAME

import(
	"strings"
)

// Global variable containing version information from
// gogetvers.
var $VARNAME = $TYPE_NAME{$VERSION,[]struct{
	Name string
	Version string
} $DEPENDENCIES}

// Contains version information for package and its dependencies.
type $TYPE_NAME struct {
	Version string
	Dependencies []struct{
		Name string
		Version string
	}
}

// Returns the version for the package.
func (vt $TYPE_NAME) GetVersion(binaryName string) string {
	return binaryName + " version " + vt.Version
}

// Returns the version for the package and all of its dependencies.
func (vt $TYPE_NAME) GetVersionVerbose(binaryName string) string {
	v := vt.GetVersion(binaryName)
	deps := []string{}
	for _,dep:=range vt.Dependencies {
		deps = append(deps,dep.Name +" version " +dep.Version)
	}
	if len(deps)>0{
		v = v + "\n    " + strings.Join(deps,"\n    ")
	}
	return v
}
`
