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
	url := fmt.Sprintf("nats://%v:%v", BrokerHost, BrokerPort)
	timeout, err := strconv.ParseInt(BrokerTimeout.Get(), 0, 64)
	if err != nil {
		return errors.Wrap(err, "Failed to convert BrokerTimeout to int")
	}
	timeoutDuration := time.Duration(timeout)
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
	commit := &schema.Commit{}
	err := proto.Unmarshal(msg.Data, commit)
	if err != nil {
		err = errors.Wrap(err, "Failed to subscribe to BrokerCommitSubject")
		log.Printf("%+v", err)
		return
	}
	RunPipeline(*commit)
}
