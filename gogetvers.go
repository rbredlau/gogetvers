package gogetvers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type GoGetVers struct {
	Path        string        // Working.Path of gogetvers.
	File        string        // Path of package info file.
	PackageInfo *PackageInfo  // The package info
	Status      *StatusWriter // The status writer.
}

func NewGoGetVers(path, file string, statusWriter io.Writer) (*GoGetVers, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	rv := &GoGetVers{Path: abs, File: file}
	if statusWriter != nil {
		rv.Status = &StatusWriter{Writer: statusWriter}
	}
	return rv, nil
}

// Use package name from manifest file if packageName is empty string.
func (g *GoGetVers) Const(outputFile, packageName string) error {
	if g == nil {
		return errors.New("nil receiver")
	}
	g.Status.Printf("Creating constant file from manifest @ %v\n", g.File)
	g.Status.Printf("Output location @ %v\n", g.Path)
	//
	var err error
	g.PackageInfo, err = LoadPackageInfoFile(g.File)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	g.Status.Writeln("Load manifest successful.")
	//
	if packageName == "" {
		packageName = filepath.Base(g.PackageInfo.PackageDir)
	}
	template := strings.Replace(version_template, "$PACKAGE_NAME", packageName, -1)
	template = strings.Replace(template, "$VARNAME", "VersionInfo", -1)
	template = strings.Replace(template, "$TYPE_NAME", "VersionInfoType", -1)
	template = strings.Replace(template, "$VERSION", fmt.Sprintf("\"%v\"", g.PackageInfo.Git.Describe), -1)
	deps := []string{}
	for _, git := range g.PackageInfo.getGits() {
		deps = append(deps, fmt.Sprintf("{\"%v\",\"%v\"}", git.HomeDir, git.Describe))
	}
	depsString := fmt.Sprintf("{%v}", strings.Join(deps, ",\n"))
	template = strings.Replace(template, "$DEPENDENCIES", depsString, -1)
	//
	fw, err := os.Create(outputFile)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	defer fw.Close()
	wrote, err := fw.WriteString(template)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	if wrote != len(template) {
		err = errors.New(fmt.Sprintf("partial file write @ %v", outputFile))
		g.Status.Error(err)
		return err
	}
	//
	cmd := NewCommandGoFmt(filepath.Base(outputFile))
	err = cmd.Exec(filepath.Dir(outputFile))
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	return nil
}

func (g *GoGetVers) Checkout() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	g.Status.Printf("Attempting to checkout manifest @ %v\n", g.File)
	g.Status.Printf("Output location @ %v\n", g.Path)
	//
	var err error
	g.PackageInfo, err = LoadPackageInfoFile(g.File)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	g.Status.Writeln("Load manifest successful.")
	//
	if !IsDir(g.Path) {
		return errors.New(fmt.Sprintf("not a path @ %v", g.Path))
	}
	g.PackageInfo.SetPathPrefix(g.Path)
	// none of g.PackageInfo.gits can have local modifications
	mods, nomods, dne, err := g.PackageInfo.getGitsLocalModsStatus()
	if err != nil {
		g.Status.Error(err)
		return err
	}
	if mods.Len() > 0 {
		g.PackageInfo.StripPathPrefix(g.Path)
		err = errors.New(fmt.Sprintf("The following gits have local modifications: %v", strings.Join(mods.Names(), ", ")))
		g.Status.Error(err)
		return err
	}
	// Checkout gits with nomods
	for _, git := range nomods {
		g.Status.Printf("checkout %v\n", git.HomeDir)
		git.Checkout()
	}
	// Clone non-existing gis
	for _, git := range dne {
		g.Status.Printf("cloning %v\n", git.HomeDir)
		git.Clone(true)
		git.Checkout()
	}
	//
	return nil
}

func (g *GoGetVers) Rebuild() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	g.Status.Printf("Attempting to rebuild manifest @ %v\n", g.File)
	g.Status.Printf("Output location @ %v\n", g.Path)
	//
	var err error
	g.PackageInfo, err = LoadPackageInfoFile(g.File)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	g.Status.Writeln("Load manifest successful.")
	//
	if !IsDir(g.Path) {
		return errors.New(fmt.Sprintf("not a path @ %v", g.Path))
	}
	g.PackageInfo.SetPathPrefix(g.Path)
	// Rebuild requires that all gits do not exist.
	exist, dne := g.PackageInfo.getGitsDiskStatus()
	if exist.Len() > 0 {
		g.PackageInfo.StripPathPrefix(g.Path)
		err = errors.New(fmt.Sprintf("The following gits already exist on disk: %v", strings.Join(exist.Names(), ", ")))
		g.Status.Error(err)
		return err
	}
	// Clone gits that do not exist.
	for _, git := range dne {
		g.Status.Printf("cloning %v\n", git.HomeDir)
		git.Clone(true)
		git.Checkout()
	}
	//
	return nil
}

func (g *GoGetVers) Release(gofile, packageName, tag, message string) error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	//
	if tag == "" {
		return errors.New("tag is empty")
	}
	// Get package information because we need to check for local modifications.
	g.PackageInfo, err = getPackageInfo(g.Path, nil)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	gits := g.PackageInfo.getGits()
	gitsWMods := []string{}
	for _, git := range gits {
		if git.Status != "" {
			gitsWMods = append(gitsWMods, git.HomeDir)
		}
	}
	if len(gitsWMods) > 0 {
		err := errors.New("The following repositories have local modifications: " + strings.Join(gitsWMods, ", "))
		g.Status.Error(err)
		return err

	}
	//
	if message == "" {
		message = tag
	}
	gittag := NewCommandGitTagAnnotated(tag, message)
	err = gittag.Exec(g.Path)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	gittagpush := NewCommandGitTagPush(tag, "origin")
	err = gittagpush.Exec(g.Path)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	err = g.Make()
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	err = g.Const(gofile, packageName)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	gitadd := NewCommand("git", "add", ".")
	gitadd.Exec(g.Path)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	gitcommit := NewCommand("git", "commit", "-m", message)
	gitcommit.Exec(g.Path)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	return nil
}

func (g *GoGetVers) Tag(tag string) error {
	if g == nil {
		return errors.New("nil receiver")
	}
	//
	if tag == "" {
		return errors.New("tag is empty")
	}
	//
	var err error
	//
	gittagdel := NewCommandGitTagDelete(tag)
	gittagdel.Exec(g.Path)
	//
	gittag := NewCommandGitTag(tag)
	err = gittag.Exec(g.Path)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	return g.Make()
}

func (g *GoGetVers) Make() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	g.PackageInfo, err = getPackageInfo(g.Path, g.Status)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	//
	g.PackageInfo.StripPathPrefix(g.PackageInfo.RootDir)
	g.Status.Writeln(g.PackageInfo.getSummary())
	//
	gits := g.PackageInfo.getGits()
	gitsWMods := []string{}
	for _, git := range gits {
		if git.Status != "" {
			gitsWMods = append(gitsWMods, git.HomeDir)
		}
	}
	if len(gitsWMods) > 0 {
		g.Status.Warning("The following dependencies have local modifications.")
		g.Status.Indent()
		g.Status.Writeln(strings.Join(gitsWMods, ", "))
		g.Status.Writeln("")
		g.Status.Outdent()
	}
	//
	g.Status.Printf("Writing output to %v\n", g.File)
	g.Status.Indent()
	fw, err := os.Create(g.File)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	defer fw.Close()
	//
	enc := json.NewEncoder(fw)
	err = enc.Encode(g.PackageInfo)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	g.Status.Writeln("done")
	g.Status.Outdent()
	g.Status.Writeln("")
	//
	return nil
}

func (g *GoGetVers) Print() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	g.PackageInfo, err = LoadPackageInfoFile(g.File)
	if err != nil {
		g.Status.Error(err)
		return err
	}
	g.Status.Writeln(g.PackageInfo.getSummary())
	return nil
}
