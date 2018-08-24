package main

import (
	"github.com/slamdev/micro-ci/services/pipeline-runner/internal"
	"log"
)

func main() {
	var err error
	err = internal.ListenBroker()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	err = internal.StartServer()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
