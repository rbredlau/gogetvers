package gogetvers

type PackageInfo struct {
	PackageDir  string                     // Package source directory; absolute path.
	GitDir      string                     // Path to .git for package.
	Git         *Git                       // Git info for package.
	Deps        []string                   // List of package dependencies.
	GoSrcDir    string                     // Absolute path of Go src that contains SourceDir.
	DepInfo     map[string]*DependencyInfo // Map of dependency info.
	GitDirs     map[string][]*DependencyInfo
	Gits        map[string]*Git
	Untrackable map[string]*DependencyInfo
	//
	*pathsComposite
}

func newPackageInfo() *PackageInfo {
	rv := &PackageInfo{Deps: []string{},
		DepInfo:     make(map[string]*DependencyInfo),
		GitDirs:     make(map[string][]*DependencyInfo),
		Gits:        make(map[string]*Git),
		Untrackable: make(map[string]*DependencyInfo)}
	rv.pathsComposite = newPathsComposite(&rv.PackageDir, &rv.GitDir)
	return rv
}

// If prefix is empty string then p.GoSrcDir is used instead.
func (p *PackageInfo) StripPathPrefix(prefix string) {
	if p == nil {
		return
	}
	if prefix == "" {
		prefix = p.GoSrcDir
	}
	p.pathsComposite.StripPathPrefix(prefix)
	if p.Git != nil {
		p.Git.StripPathPrefix(prefix)
	}
	for _, v := range p.DepInfo {
		v.StripPathPrefix(prefix)
	}
	for _, v := range p.Gits {
		v.StripPathPrefix(prefix)
	}
}

func (p *PackageInfo) PathsToSlash() {
	if p == nil {
		return
	}
	p.pathsComposite.PathsToSlash()
	// Do the GoSrcDir too
	pc := newPathsComposite(&p.GoSrcDir)
	pc.PathsToSlash()
	if p.Git != nil {
		p.Git.PathsToSlash()
	}
	for _, v := range p.DepInfo {
		v.PathsToSlash()
	}
	for _, v := range p.Gits {
		v.PathsToSlash()
	}
}

func (p *PackageInfo) PathsFromSlash() {
	if p == nil {
		return
	}
	p.pathsComposite.PathsFromSlash()
	// Do the GoSrcDir too
	pc := newPathsComposite(&p.GoSrcDir)
	pc.PathsFromSlash()
	if p.Git != nil {
		p.Git.PathsFromSlash()
	}
	for _, v := range p.DepInfo {
		v.PathsFromSlash()
	}
	for _, v := range p.Gits {
		v.PathsFromSlash()
	}
}
