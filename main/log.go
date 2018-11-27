package main

import (
	"os"

	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("go-multi-assistant")
)

// initializeLogger initializes the log engine
func initializeLogger(errorLogFilepath string) {
	// create/open a file to log all errors
	flog, err := os.OpenFile(errorLogFilepath, os.O_RDWR|os.O_CREATE, 0755)
	if err == nil {
		defer flog.Close()
	}

	// create two backends, one for a file and the second for os.Stderr.
	backend1 := logging.NewLogBackend(flog, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
	)
	// add the same format to both loggers
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	// Only errors and more severe messages should be sent to backend1 (the file)
	backend1Leveled := logging.SetBackend(backend1Formatter)
	backend1Leveled.SetLevel(logging.ERROR, "")
	// Set the backends to be used.
	if err == nil {
		logging.SetBackend(backend1Leveled, backend2Formatter)
	} else {
		logging.SetBackend(backend2Formatter)
	}
}
