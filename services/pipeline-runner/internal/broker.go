package internal

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/slamdev/micro-ci/etc/schema"
	"log"
	"strconv"
	"time"
)

func ListenBroker() error {
	url := fmt.Sprintf("nats://%v:%v", BrokerHost.Get(), BrokerPort.Get())
	timeout, err := strconv.Atoi(BrokerTimeout.Get())
	if err != nil {
		return errors.Wrap(err, "Failed to convert BrokerTimeout to int")
	}
	timeoutDuration := time.Duration(timeout) * time.Second
	nc, err := nats.Connect(url, nats.ReconnectWait(timeoutDuration), nats.Timeout(timeoutDuration))
	if err != nil {
		return errors.Wrap(err, "Failed to connect to nats server")
	}
	_, err = nc.Subscribe(BrokerCommitSubject.Get(), handleCommitMessage)
	if err != nil {
		return errors.Wrap(err, "Failed to subscribe to BrokerCommitSubject")
	}
	return nil
}

func handleCommitMessage(msg *nats.Msg) {
	commit, err := convert(msg)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	pipeline, err := FetchPipeline(commit)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
	err = RunPipeline(pipeline, commit)
	if err != nil {
		log.Printf("%+v", err)
		return
	}
}

func convert(msg *nats.Msg) (schema.Commit, error) {
	commit := &schema.Commit{}
	err := proto.Unmarshal(msg.Data, commit)
	if err != nil {
		return *commit, errors.Wrap(err, "Failed to unmarshal commit message")
	}
	return *commit, nil
}
