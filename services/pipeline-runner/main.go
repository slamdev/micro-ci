package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/slamdev/micro-ci/etc/schema"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	batch "k8s.io/api/batch/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	log.Print("Starting [pipeline-runner] service")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "It's Alive!")
	})
	listen()
	send()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func send() {
	nc, err := nats.Connect(
		"nats://"+os.Getenv("MESSAGE_BROKER_SVC_SERVICE_HOST")+":"+os.Getenv("MESSAGE_BROKER_SVC_SERVICE_PORT_CLIENT"),
		nats.ReconnectWait(15*time.Second),
		nats.Timeout(15*time.Second))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(nc)
	commit := &schema.Commit{
		Author:   "slamdev",
		Branch:   "master",
		Revision: "758e01c07c119aeadb5dc31f7aadac69ece69ebd",
	}
	data, err := proto.Marshal(commit)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = nc.Publish("commit", data)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func listen() {
	nc, err := nats.Connect(
		"nats://"+os.Getenv("MESSAGE_BROKER_SVC_SERVICE_HOST")+":"+os.Getenv("MESSAGE_BROKER_SVC_SERVICE_PORT_CLIENT"),
		nats.ReconnectWait(15*time.Second),
		nats.Timeout(15*time.Second))
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(nc)
	sub, err := nc.Subscribe("commit", func(msg *nats.Msg) {
		commit := &schema.Commit{}
		err := proto.Unmarshal(msg.Data, commit)
		if err != nil {
			log.Fatal(err)
			return
		}
		runPipeline(commit)
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(sub)
}

func runPipeline(commit *schema.Commit) {
	repoUrl := "https://github.com/slamdev/micro-ci.git"
	cloneDir := "build/clone"
	execCommand("git clone --branch master "+repoUrl+" "+cloneDir, "")
	execCommand("git checkout "+commit.Branch, cloneDir)
	execCommand("git reset --hard "+commit.Revision, cloneDir)
	execCommand("git merge --no-ff --no-commit origin/master", cloneDir)
	content, err := ioutil.ReadFile(cloneDir + ".pipeline.yaml")
	if err != nil {
		log.Fatal(err)
	}
	pipeline := &Pipeline{}
	err = yaml.UnmarshalStrict(content, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	for _, job := range pipeline.Jobs {
		runJob(job)
	}
}

func runJob(job Job) {
	log.Println(job)
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("HOME") + "/.kube/config"
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	jobTemplate := &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name: "sample",
		},
		Spec: batch.JobSpec{},
	}
	log.Println(jobTemplate)
	_, err = clientset.BatchV1().Jobs("micro-ci").Create(jobTemplate)
	if err != nil {
		log.Fatal(err)
	}
	jobList, err := clientset.BatchV1().Jobs("micro-ci").List(meta.ListOptions{})
	log.Println(jobList.Items)
}

func execCommand(command string, dir string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

type Pipeline struct {
	Jobs []Job
}

type Job struct {
	Name string
	Spec map[interface{}]interface{}
}
