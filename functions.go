package gogetvers

import (
	fs "broadlux/fileSystem"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Opens the input file and decodes the manifest.
func LoadManifest(inputFile string) (*PackageSummary, error) {
	if !fs.IsFile(inputFile) {
		return nil, errors.New(fmt.Sprintf("Not a file @ %v", inputFile))
	}
	fr, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	//
	dec := json.NewDecoder(fr)
	summary := &PackageSummary{}
	err = dec.Decode(summary)
	if err != nil {
		return nil, err
	}
	return summary, nil
}
