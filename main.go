package main

import (
	"flag"
	"github.com/santihernandezc/devnull/service"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	port       = flag.String("p", "8080", "Port to listen on")
	target     = flag.String("t", "", "Target (URL) to forward requests to")
	output     = flag.String("o", "", "Output file for logs")
	wait       = flag.Duration("w", 0, "Minimum wait time before HTTP response")
	statusCode = flag.Int("status", 200, "Status code used in responses if no target is configured")

	verbose = flag.Bool("v", false, "Enable verbose logging")
)

func init() {
	flag.Parse()
}

func main() {
	//	fmt.Println(`
	//██████╗░███████╗██╗░░░██╗███╗░░██╗██╗░░░██╗██╗░░░░░██╗░░░░░
	//██╔══██╗██╔════╝██║░░░██║████╗░██║██║░░░██║██║░░░░░██║░░░░░
	//██║░░██║█████╗░░╚██╗░██╔╝██╔██╗██║██║░░░██║██║░░░░░██║░░░░░
	//██║░░██║██╔══╝░░░╚████╔╝░██║╚████║██║░░░██║██║░░░░░██║░░░░░
	//██████╔╝███████╗░░╚██╔╝░░██║░╚███║╚██████╔╝███████╗███████╗
	//╚═════╝░╚══════╝░░░╚═╝░░░╚═╝░░╚══╝░╚═════╝░╚══════╝╚══════╝
	//
	//It does literally nothing`)

	writers := []io.Writer{os.Stdout}
	if *output != "" {
		path := filepath.Join(".", *output)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.Fatalf("msg=%q err=%q\n", "Error creating output directory", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("msg=%q err=%s\n", "Error opening output file", err)
		}
		log.Printf("msg=%q\n", "Writing logs to "+path)
		writers = append(writers, f)
	}

	log.Printf("msg=%q\n", "Listening on port "+*port)

	log := log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime|log.Lmicroseconds)
	svc := service.New(log, *target, *verbose, *statusCode, *wait)

	http.HandleFunc("/", svc.Handler)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("msg=%q err=%q\n", "Error from ListenAndServe", err)
	}
}
