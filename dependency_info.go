package gogetvers

type DependencyInfo struct {
	IsGo   bool   // True if not in GoSrcDir as a path.
	IsGit  bool   // True if .git info was found.
	Name   string // Name according to: go list -f {{.Deps}} from the parent package.
	DepDir string // Path to dependency.
	GitDir string // The .git directory.
	//
	*pathsComposite
}

func newDependencyInfo(isGo, isGit bool, name, dir string) *DependencyInfo {
	rv := &DependencyInfo{IsGo: isGo, IsGit: isGit, Name: name, DepDir: dir}
	rv.pathsComposite = newPathsComposite(&rv.DepDir, &rv.GitDir)
	return rv
}
