package main

import (
	"flag"
	"github.com/santihernandezc/devnull/service"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"rsc.io/getopt"
	"time"
)

var (
	debug      = flag.Bool("debug", false, "Enable debug logging")
	output     = flag.String("output", "", "Output file for logs")
	port       = flag.String("port", "8080", "Port to listen on")
	statusCode = flag.Int("status-code", 200, "Status code used in responses if no target is configured")
	target     = flag.String("target", "", "Target (URL) to forward requests to")
	timeout    = flag.Duration("timeout", 30*time.Second, "Timeout for the HTTP client")
	verbose    = flag.Bool("verbose", false, "Log extra details about the request (headers, request body)")
	wait       = flag.Duration("wait", 0, "Minimum wait time before HTTP response")
)

func init() {
	getopt.Aliases(
		"d", "debug",
		"o", "output",
		"p", "port",
		"s", "status-code",
		"t", "target",
		"T", "timeout",
		"v", "verbose",
		"w", "wait",
	)
	getopt.Parse()
}

func main() {
	log := logrus.New()
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	}

	if *output != "" {
		path := filepath.Join(".", *output)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.WithError(err).Error("msg", "Error creating output directory")
			return
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.WithError(err).Error("msg", "Error opening output file")
			return
		}

		log.WithField("path", path).Debug("Writing logs to file")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	svc, err := service.New(log, *target, *verbose, *statusCode, *timeout, *wait)
	if err != nil {
		log.WithError(err).Fatalf("Error creating service")
	}

	http.HandleFunc("/", svc.Handler)

	log.WithField("port", *port).Info("Starting server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatalf("Error from ListenAndServe")
	}
}
