package internal

import (
	"github.com/pkg/errors"
	"github.com/slamdev/micro-ci/etc/schema"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
)

func RunPipeline(commit schema.Commit) error {
	pipeline, err := fetchPipeline(commit)
	if err != nil {
		return err
	}
	for _, job := range pipeline.Jobs {
		err = runJob(job)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchPipeline(commit schema.Commit) (Pipeline, error) {
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

func runJob(job Job) error {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		err = errors.Wrap(err, "Failed to get k8s config from cluster; local config will be used")
		log.Printf("%+v", err)
		kubeconfig := os.Getenv("HOME") + "/.kube/config"
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return errors.Wrap(err, "Failed to get local k8s config")
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "Failed to connect to cluster")
	}
	jobTemplate := &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      "sample",
			Namespace: "micro-ci",
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				Spec: job.Spec,
			},
		},
	}
	jobTemplate.Spec.Template.Spec.RestartPolicy = core.RestartPolicyNever
	instance, err := clientset.BatchV1().Jobs("").Create(jobTemplate)
	if err != nil {
		return errors.Wrap(err, "Failed to create k8s Job resource")
	}
	log.Println(instance)
	return nil
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
