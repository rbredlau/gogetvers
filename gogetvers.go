package gogetvers

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type GoGetVers struct {
	path        string        // Working path of gogetvers.
	file        string        // Path of package info file.
	packageInfo *PackageInfo  // The package info
	status      *StatusWriter // The status writer.
}

func NewGoGetVers(path, file string, statusWriter io.Writer) (*GoGetVers, error) {
	rv := &GoGetVers{path: path, file: file}
	if statusWriter != nil {
		rv.status = &StatusWriter{Writer: statusWriter}
	}
	return rv, nil
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
