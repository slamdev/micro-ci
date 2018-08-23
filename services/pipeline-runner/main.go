package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/slamdev/micro-ci/etc/schema"
	"log"
	"net/http"
	"os"
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
		Revision: "123",
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
		log.Println(commit)
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(sub)
}
