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
)

var (
	port       = flag.String("port", "8080", "Port to listen on")
	target     = flag.String("target", "", "Target (URL) to forward requests to")
	output     = flag.String("output", "", "Output file for logs")
	wait       = flag.Duration("wait", 0, "Minimum wait time before HTTP response")
	statusCode = flag.Int("status-code", 200, "Status code used in responses if no target is configured")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func init() {
	getopt.Aliases(
		"p", "port",
		"t", "target",
		"o", "output",
		"w", "wait",
		"s", "status-code",
		"v", "verbose",
	)
	getopt.Parse()
}

func main() {
	log := logrus.New()
	if *verbose {
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

	svc := service.New(log, *target, *verbose, *statusCode, *wait)

	http.HandleFunc("/", svc.Handler)

	log.WithField("port", *port).Info("Starting server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatalf("Error from ListenAndServe")
	}
}
