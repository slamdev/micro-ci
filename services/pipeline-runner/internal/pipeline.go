package internal

import (
	"github.com/pkg/errors"
	"github.com/slamdev/micro-ci/etc/schema"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	core "k8s.io/api/core/v1"
	"os"
	"os/exec"
)

func FetchPipeline(commit schema.Commit) (Pipeline, error) {
	cloneDir, err := ioutil.TempDir("", "clone")
	if err != nil {
		return Pipeline{}, errors.Wrap(err, "Failed to create temp dir to clone repo")
	}
	defer os.RemoveAll(cloneDir)
	err = execCommand("git clone --branch master "+RepoUrl.Get()+" "+cloneDir, "")
	if err != nil {
		return Pipeline{}, err
	}
	err = execCommand("git checkout "+commit.Branch, cloneDir)
	if err != nil {
		return Pipeline{}, err
	}
	err = execCommand("git reset --hard "+commit.Revision, cloneDir)
	if err != nil {
		return Pipeline{}, err
	}
	err = execCommand("git merge --no-ff --no-commit origin/master", cloneDir)
	if err != nil {
		return Pipeline{}, err
	}
	content, err := ioutil.ReadFile(cloneDir + "/.pipeline.yaml")
	if err != nil {
		return Pipeline{}, errors.Wrap(err, "Failed to create read pipeline file")
	}
	pipeline := &Pipeline{}
	err = yaml.UnmarshalStrict(content, pipeline)
	if err != nil {
		return Pipeline{}, errors.Wrap(err, "Failed to convert yaml to Pipeline")
	}
	return *pipeline, nil
}

func execCommand(command string, dir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "Failed to execute shell command")
	}
	return nil
}

type Pipeline struct {
	Jobs []Job
}

type Job struct {
	Name string
	Spec core.PodSpec
}
