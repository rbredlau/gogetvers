package gogetvers

import (
	fs "broadlux/fileSystem"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type GoGetVers struct {
	path        string        // Working path of gogetvers.
	file        string        // Path of package info file.
	packageInfo *PackageInfo  // The package info
	status      *StatusWriter // The status writer.
}

func NewGoGetVers(path, file string, statusWriter io.Writer) (*GoGetVers, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	rv := &GoGetVers{path: abs, file: file}
	if statusWriter != nil {
		rv.status = &StatusWriter{Writer: statusWriter}
	}
	return rv, nil
}

func (g *GoGetVers) Checkout() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	g.status.Printf("Attempting to checkout manifest @ %v\n", g.file)
	g.status.Printf("Output location @ %v\n", g.path)
	//
	var err error
	g.packageInfo, err = loadPackageInfoFile(g.file)
	if err != nil {
		g.status.Error(err)
		return err
	}
	g.status.Writeln("Load manifest successful.")
	//
	if !fs.IsDir(g.path) {
		return errors.New(fmt.Sprintf("not a path @ %v", g.path))
	}
	g.packageInfo.SetPathPrefix(g.path)
	// none of g.packageInfo.gits can have local modifications
	mods, nomods, dne, err := g.packageInfo.getGitsLocalModsStatus()
	if err != nil {
		g.status.Error(err)
		return err
	}
	if mods.Len() > 0 {
		g.packageInfo.StripPathPrefix(g.path)
		err = errors.New(fmt.Sprintf("The following gits have local modifications: %v", strings.Join(mods.Names(), ", ")))
		g.status.Error(err)
		return err
	}
	// Checkout gits with nomods
	for _, git := range nomods {
		g.status.Printf("checkout %v\n", git.HomeDir)
		git.Checkout()
	}
	// Clone non-existing gis
	for _, git := range dne {
		g.status.Printf("cloning %v\n", git.HomeDir)
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
	g.status.Printf("Attempting to rebuild manifest @ %v\n", g.file)
	g.status.Printf("Output location @ %v\n", g.path)
	//
	var err error
	g.packageInfo, err = loadPackageInfoFile(g.file)
	if err != nil {
		g.status.Error(err)
		return err
	}
	g.status.Writeln("Load manifest successful.")
	//
	if !fs.IsDir(g.path) {
		return errors.New(fmt.Sprintf("not a path @ %v", g.path))
	}
	g.packageInfo.SetPathPrefix(g.path)
	// Rebuild requires that all gits do not exist.
	exist, dne := g.packageInfo.getGitsDiskStatus()
	if exist.Len() > 0 {
		g.packageInfo.StripPathPrefix(g.path)
		err = errors.New(fmt.Sprintf("The following gits already exist on disk: %v", strings.Join(exist.Names(), ", ")))
		g.status.Error(err)
		return err
	}
	// Clone gits that do not exist.
	for _, git := range dne {
		g.status.Printf("cloning %v\n", git.HomeDir)
		git.Clone(true)
		git.Checkout()
	}
	//
	return nil
}

func (g *GoGetVers) Make() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	g.packageInfo, err = getPackageInfo(g.path, g.status)
	if err != nil {
		g.status.Error(err)
		return err
	}
	//
	g.packageInfo.StripPathPrefix(g.packageInfo.RootDir)
	g.status.Writeln(g.packageInfo.getSummary())
	//
	g.status.Printf("Writing output to %v\n", g.file)
	g.status.Indent()
	fw, err := os.Create(g.file)
	if err != nil {
		g.status.Error(err)
		return err
	}
	defer fw.Close()
	//
	enc := json.NewEncoder(fw)
	err = enc.Encode(g.packageInfo)
	if err != nil {
		g.status.Error(err)
		return err
	}
	g.status.Writeln("done")
	g.status.Outdent()
	g.status.Writeln("")
	//
	return nil
}

func (g *GoGetVers) Print() error {
	if g == nil {
		return errors.New("nil receiver")
	}
	var err error
	g.packageInfo, err = loadPackageInfoFile(g.file)
	if err != nil {
		g.status.Error(err)
		return err
	}
	g.status.Writeln(g.packageInfo.getSummary())
	return nil
}