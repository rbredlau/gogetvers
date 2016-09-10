package gogetvers

const version_template = `
package $PACKAGE_NAME

import(
	"strings"
)

var $VARNAME = $TYPE_NAME{$VERSION,[]struct{
	Name string
	Version string
} $DEPENDENCIES}

type $TYPE_NAME struct {
	Version string
	Dependencies []struct{
		Name string
		Version string
	}
}

func (v $TYPE_NAME) Version(binaryName string) string {
	return binaryName + " version " + v.Version
}

func (v $TYPE_NAME) VersionVerbose(binaryName string) string {
	v := v.Version(binaryName) + "\n"
	deps := []string{}
	for _,dep:=range v.Dependencies {
		deps = append(deps,dep.Name +" version " +dep.Version)
	}
	if len(deps)>0{
		v = v + "    " + strings.Join(deps,"\n    ")
	}
	return v
}
`
