package main

import (
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

type Candidate struct {
	Name     string
	FullPath string
}

// KubeconfigFilenamePattern defines the name pattern of kubeconfig files
var KubeconfigFilenamePattern = regexp.MustCompile("^(.*)\\.(kubeconfig|config)$")

// ListKubeconfigCandidatesInDir lists all files in dir that matches KubeconfigFilenamePattern
func ListKubeconfigCandidatesInDir(dir string) ([]Candidate, error) {
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadDir error")
	}

	var files []Candidate
	for _, file := range fileInfo {
		if file.IsDir() || IsSymlink(file) {
			continue
		}

		if file.Name() == "config" {
			files = append(files, Candidate{
				Name:     file.Name(),
				FullPath: filepath.Join(dir, file.Name()),
			})
			continue
		}

		matches := KubeconfigFilenamePattern.FindStringSubmatch(file.Name())
		if len(matches) >= 2 {
			files = append(files, Candidate{
				Name:     matches[1],
				FullPath: filepath.Join(dir, file.Name()),
			})
		}
	}
	return files, nil
}
