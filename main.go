package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	port      = flag.String("port", "8080", "Port to listen on")
	forwardTo = flag.String("target", "", "URL to forward requests to")
	verbose   = flag.Bool("v", false, "Enable verbose logging")
	output    = flag.String("o", "", "Output file")
)

func init() {
	flag.Parse()
}

func main() {
	fmt.Println(`
██████╗░███████╗██╗░░░██╗███╗░░██╗██╗░░░██╗██╗░░░░░██╗░░░░░
██╔══██╗██╔════╝██║░░░██║████╗░██║██║░░░██║██║░░░░░██║░░░░░
██║░░██║█████╗░░╚██╗░██╔╝██╔██╗██║██║░░░██║██║░░░░░██║░░░░░
██║░░██║██╔══╝░░░╚████╔╝░██║╚████║██║░░░██║██║░░░░░██║░░░░░
██████╔╝███████╗░░╚██╔╝░░██║░╚███║╚██████╔╝███████╗███████╗
╚═════╝░╚══════╝░░░╚═╝░░░╚═╝░░╚══╝░╚═════╝░╚══════╝╚══════╝
	
It does literally nothing
`)

	writers := []io.Writer{os.Stdout}
	if *output != "" {
		path := filepath.Join(".", *output)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			log.Fatalf("Error creating output directory: %v\n", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Error opening output file: %v\n", err)
		}
		log.Print("Writing logs to " + path)
		writers = append(writers, f)
	}

	log := log.New(io.MultiWriter(writers...), "", log.Ldate|log.Ltime|log.Lmicroseconds)
	log.Printf("Listening on port %s\n", *port)

	target := strings.TrimSpace(*forwardTo)
	if target != "" {
		log.Printf("Forwarding requests to %s\n", target)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !*verbose {
			log.Printf("msg=\"Request received\" method=%q URI=%q\n", r.Method, r.RequestURI)
		} else {
			var headers []string
			for k, v := range r.Header {
				headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
			}

			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Printf("msg=\"Error reading body\" err=%v\n", err)
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				log.Printf("msg=\"Request received\" method=%q URI=%q headers=%q body=%q\n", r.Method, r.RequestURI, strings.Join(headers, ", "), string(body))
			} else {
				fmt.Printf("msg=\"Request received\" method=%q URI=%q headers=%q\n", r.Method, r.RequestURI, strings.Join(headers, ", "))
			}
		}

		if target != "" {
			uri, err := url.JoinPath(target, r.RequestURI)
			if err != nil {
				log.Printf("msg=\"Error joining paths\" target=%q path=%q error=%q\n", target, r.URL.Path, err)
				return
			}

			res, err := forwardRequest(r, uri)
			if err != nil {
				log.Printf("msg=\"Error forwarding request\" target=%q error=%q\n", target, err)
				return
			}

			for k, v := range res.Header {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(res.StatusCode)

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Printf("msg=\"Error reading response body\" error=%q\n", err)
			}

			defer res.Body.Close()
			if _, err := w.Write(body); err != nil {
				log.Printf("msg=\"Error writing response body\" error=%q\n", err)
			}
		}
	})

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func forwardRequest(r *http.Request, target string) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	return http.DefaultClient.Do(req)
}
