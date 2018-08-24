package internal

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"time"
)

type Config string

const (
	ServerPort          Config = "SERVER_PORT"
	BrokerHost          Config = "MESSAGE_BROKER_SVC_SERVICE_HOST"
	BrokerPort          Config = "MESSAGE_BROKER_SVC_SERVICE_PORT_CLIENT"
	BrokerTimeout       Config = "MESSAGE_BROKER_TIMEOUT"
	BrokerCommitSubject Config = "MESSAGE_BROKER_COMMIT_SUBJECT"
	RepoUrl             Config = "CI_REPO_URL"
)

var defaults = map[Config]string{
	ServerPort:          "8080",
	BrokerHost:          "localhost",
	BrokerPort:          "4222",
	BrokerTimeout:       string(15 * time.Second),
	BrokerCommitSubject: "commit",
	RepoUrl:             "https://github.com/slamdev/micro-ci.git",
}

func (config Config) Get() string {
	value, exists := os.LookupEnv(string(config))
	if !exists {
		value, exists = defaults[config]
		if !exists {
			err := errors.New("Failed to subscribe to BrokerCommitSubject")
			log.Fatalf("%+v", err)
		}
	}
	return value
}
