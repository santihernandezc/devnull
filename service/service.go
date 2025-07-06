package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type service struct {
	log        *logrus.Logger
	statusCode int
	target     string
	verbose    bool
	wait       time.Duration
}

func New(log *logrus.Logger, target string, verbose bool, statusCode int, wait time.Duration) *service {
	target = strings.TrimSpace(target)
	if target != "" {
		log.WithField("target", target).Debug("Forwarding requests to target")
	}

	if statusCode < 100 || statusCode > 999 {
		if statusCode != 0 {
			panic(fmt.Sprintf("Invalid status code %d", statusCode))
		}
		statusCode = http.StatusOK
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

	if s.target == "" {
		// No target configured, sleep and respond.
		time.Sleep(s.wait)
		if s.statusCode != 200 {
			w.WriteHeader(s.statusCode)
		}
		return
	}

	uri, err := url.JoinPath(s.target, r.RequestURI)
	if err != nil {
		s.log.WithError(err).
			WithField("target", s.target).
			WithField("path", r.URL.Path).
			Error("Error joining paths")
		return
	}

	res, err := forwardRequest(r, uri)
	if err != nil {
		s.log.WithError(err).WithField("target", s.target).Error("Error forwarding request")
		time.Sleep(s.wait)
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
		s.log.WithError(err).Error("Error reading response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()
	time.Sleep(s.wait)
	if _, err := w.Write(body); err != nil {
		s.log.WithError(err).Error("Error writing response body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *service) logRequestDetails(r *http.Request) {
	if !s.verbose {
		s.log.WithField("method", r.Method).WithField("uri", r.RequestURI).Info("Request received")
		return
	}

	var headers []string
	for k, v := range r.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}
	sort.Strings(headers)

	fields := logrus.Fields{
		"method":  r.Method,
		"uri":     r.RequestURI,
		"headers": strings.Join(headers, ", "),
	}

	if r.Method == http.MethodPatch || r.Method == http.MethodPost || r.Method == http.MethodPut {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.log.WithError(err).Error("Error reading request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		fields["body"] = string(body)
	}
	s.log.WithFields(fields).Info("Request received")
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
