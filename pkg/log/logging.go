package log

import (
	"github.com/op/go-logging"
	"os"
)

// Log is the handler for the logger
var Log = logging.MustGetLogger("vikunja")

var format = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// InitLogger initializes the global log handler
func InitLogger() {
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
}
