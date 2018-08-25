package internal

import (
	"github.com/pkg/errors"
	"github.com/slamdev/micro-ci/etc/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

func RunPipeline(pipeline Pipeline, commit schema.Commit) error {
	for _, job := range pipeline.Jobs {
		err := runJob(job, commit)
		if err != nil {
			return err
		}
	}
	return nil
}

func runJob(job Job, commit schema.Commit) error {
	clientset, err := connectToCluster()
	if err != nil {
		return err
	}
	podTemplate := CreatePodTemplate(job, commit)
	podInstance, err := clientset.CoreV1().Pods(podTemplate.Namespace).Create(&podTemplate)
	if err != nil {
		return errors.Wrap(err, "Failed to create k8s pod resource")
	}
	log.Println("Running pod", podInstance)
	return nil
}

func connectToCluster() (kubernetes.Clientset, error) {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		err = errors.Wrap(err, "Failed to get k8s config from cluster; local config will be used")
		log.Printf("%+v", err)
		kubeconfig := os.Getenv("HOME") + "/.kube/config"
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return kubernetes.Clientset{}, errors.Wrap(err, "Failed to get local k8s config")
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrap(err, "Failed to connect to cluster")
	}
	return *clientset, nil
}
