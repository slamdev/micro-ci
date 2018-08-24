package internal

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

func StartServer() error {
	http.HandleFunc("/health", handleHealth)
	err := http.ListenAndServe(":"+ServerPort.Get(), nil)
	return errors.Wrap(err, "Failed to start server")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
