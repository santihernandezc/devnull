package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/santihernandezc/devnull/client"
	"github.com/santihernandezc/devnull/service"
	"github.com/sirupsen/logrus"
	"rsc.io/getopt"
)

var (
	debug       = flag.Bool("debug", false, "Enable debug logging")
	metricsPort = flag.String("metrics-port", "8081", "Port to serve metrics on")
	output      = flag.String("output", "", "Output file for logs")
	port        = flag.String("port", "8080", "Port to listen on")
	statusCode  = flag.Int("status-code", 200, "Status code used in responses if no target is configured")
	target      = flag.String("target", "", "Target (URL) to forward requests to")
	timeout     = flag.Duration("timeout", 30*time.Second, "Timeout for the HTTP client, 0 = no timeout")
	verbose     = flag.Bool("verbose", false, "Log extra details about the request (headers, request body)")
	wait        = flag.Duration("wait", 0, "Minimum wait time before HTTP response")
)

func init() {
	getopt.Aliases(
		"d", "debug",
		"m", "metrics-port",
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

	ic := client.NewInstrumented(prometheus.DefaultRegisterer, *timeout)
	svc, err := service.New(log, ic, *target, *verbose, *statusCode, *wait)
	if err != nil {
		log.WithError(err).Fatalf("Error creating service")
	}

	http.HandleFunc("/", svc.Handler)

	// Start the metrics server.
	go func() {
		m := http.NewServeMux()
		m.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+*metricsPort, m); err != nil {
			log.WithError(err).Fatalf("Error from metrics server's ListenAndServe")
		}
	}()

	// Start the main server.
	log.WithField("port", *port).Info("Starting server")
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.WithError(err).Fatalf("Error from ListenAndServe")
	}
}
