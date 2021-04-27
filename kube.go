package main

import (
	"github.com/pkg/errors"
	"github.com/spongeprojects/magicconch"
	"io/ioutil"
	"os/user"
	"path"
	"regexp"
)

type Candidate struct {
	Name     string
	FullPath string
}

// KubeconfigPattern defines the name pattern of kubeconfig files
var KubeconfigPattern = regexp.MustCompile("^(.*)\\.kubeconfig$")

// ListKubeconfigCandidatesInDir lists all files in dir that matches KubeconfigPattern
func ListKubeconfigCandidatesInDir(dir string) ([]Candidate, error) {
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadDir")
	}

	var files []Candidate
	for _, file := range fileInfo {
		if !file.IsDir() {
			matches := KubeconfigPattern.FindStringSubmatch(file.Name())
			if len(matches) == 2 {
				files = append(files, Candidate{
					Name:     matches[1],
					FullPath: path.Join(dir, file.Name()),
				})
			}
		}
	}
	return files, nil
}

// KubeDir gets the .kube dir
func KubeDir() string {
	currentUser, err := user.Current()
	magicconch.Must(err)
	return path.Join(currentUser.HomeDir, ".kube")
}

// ListKubeconfigCandidates lists all files in default kube dir that matches KubeconfigPattern
func ListKubeconfigCandidates() ([]Candidate, error) {
	return ListKubeconfigCandidatesInDir(KubeDir())
}
