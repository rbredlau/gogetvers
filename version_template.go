package gogetvers

const version_template = `
package $PACKAGE_NAME

const $CONSTANT_NAME = &$TYPE_NAME{$VERSION,$DEPENDENCIES}

type $TYPE_NAME struct {
	Version string
	Dependencies []struct{
		Name string
		Version string
	}
}
`
