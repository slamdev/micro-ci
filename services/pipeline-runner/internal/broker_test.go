package internal

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/slamdev/micro-ci/etc/schema"
	"strconv"
	"testing"
	"time"
)

func TestListenBroker(t *testing.T) {
	err := ListenBroker()
	if err != nil {
		t.Fatalf("Failed with error %+v", err)
	}
	sendMessage(&schema.Commit{
		Author:   "slamdev",
		Branch:   "master",
		Revision: "b706fc1675e364f5cfd054b662560c7468f843b8",
	})
	time.Sleep(15 * time.Second)
}

func sendMessage(commit *schema.Commit) {
	url := fmt.Sprintf("nats://%v:%v", BrokerHost.Get(), BrokerPort.Get())
	timeout, _ := strconv.Atoi(BrokerTimeout.Get())
	timeoutDuration := time.Duration(timeout) * time.Second
	nc, _ := nats.Connect(url, nats.ReconnectWait(timeoutDuration), nats.Timeout(timeoutDuration))
	data, _ := proto.Marshal(commit)
	nc.Publish(BrokerCommitSubject.Get(), data)
}
