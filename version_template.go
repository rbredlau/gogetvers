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

func (v $TYPE_NAME) GetVersion(binaryName string) string {
	return binaryName + " version " + v.Version
}

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
