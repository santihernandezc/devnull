package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type service struct {
	log        *log.Logger
	statusCode int
	target     string
	verbose    bool
	wait       time.Duration
}

func New(log *log.Logger, target string, verbose bool, statusCode int, wait time.Duration) *service {
	target = strings.TrimSpace(target)
	if target != "" {
		log.Printf("msg=%q target=%q\n", "Forwarding requests to target", target)
	}

	return &service{
		log:        log,
		target:     target,
		verbose:    verbose,
		statusCode: statusCode,
		wait:       wait,
	}
}

func (s *service) Handler(w http.ResponseWriter, r *http.Request) {
	s.logRequestDetails(r)

	if s.target != "" {
		uri, err := url.JoinPath(s.target, r.RequestURI)
		if err != nil {
			s.log.Printf("msg=%q target=%q path=%q error=%q\n", "Error joining paths", s.target, r.URL.Path, err)
			return
		}

		res, err := forwardRequest(r, uri)
		if err != nil {
			s.log.Printf("msg=%q target=%q error=%q\n", "Error forwarding request", s.target, err)
			w.WriteHeader(http.StatusBadGateway)
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
			s.log.Printf("msg=%q error=%q\n", "Error reading response body", err)
		}

		defer res.Body.Close()
		time.Sleep(s.wait)
		if _, err := w.Write(body); err != nil {
			s.log.Printf("msg=%q error=%q\n", "Error writing response body", err)
		}
		return
	}
	time.Sleep(s.wait)
	if s.statusCode != 200 {
		w.WriteHeader(s.statusCode)
	}
}

func (s *service) logRequestDetails(r *http.Request) {
	if !s.verbose {
		s.log.Printf("msg=%q method=%q url=%q\n", "Request received", r.Method, r.RequestURI)
		return
	}

	var headers []string
	for k, v := range r.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}

	switch r.Method {
	case http.MethodPatch, http.MethodPost, http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.log.Printf("msg=%q err=%q\n", "Error reading request body", err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		s.log.Printf("msg=%q method=%q URI=%q headers=%q body=%q\n", "Request received", r.Method, r.RequestURI, strings.Join(headers, ", "), string(body))
	default:
		fmt.Printf("msg=%q method=%q URI=%q headers=%q\n", "Request received", r.Method, r.RequestURI, strings.Join(headers, ", "))
	}
}

func forwardRequest(r *http.Request, target string) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	// TODO: Timeouts, use custom http.Client.
	return http.DefaultClient.Do(req)
}
