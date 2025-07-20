package service

import (
	"bytes"
	"encoding/json"
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
	client     *http.Client
	log        *logrus.Logger
	statusCode int
	target     string
	verbose    bool
	wait       time.Duration
}

type response struct {
	Message string `json:"message"`
}

func New(log *logrus.Logger, c *http.Client, target string, verbose bool, statusCode int, wait time.Duration) (*service, error) {
	target = strings.TrimSpace(target)
	if target != "" {
		log.WithField("target", target).Debug("Forwarding requests to target")
	}

	if statusCode < 100 || statusCode > 999 {
		if statusCode != 0 {
			return nil, fmt.Errorf("invalid status code: %d", statusCode)
		}
		statusCode = http.StatusOK
	}

	return &service{
		client:     c,
		log:        log,
		target:     target,
		verbose:    verbose,
		statusCode: statusCode,
		wait:       wait,
	}, nil
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
		s.writeResponse(w, http.StatusInternalServerError, "Could not build URL for the target")
		return
	}

	res, err := s.forwardRequest(r, uri)
	if err != nil {
		s.log.WithError(err).WithField("target", s.target).Error("Error forwarding request")
		time.Sleep(s.wait)
		s.writeResponse(w, http.StatusBadGateway, "Could not forward request to the target")
		return
	}

	for k, v := range res.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.WithError(err).Error("Error reading response body")
		s.writeResponse(w, http.StatusInternalServerError, "Could not read the response body from the target")
		return
	}
	defer res.Body.Close()

	time.Sleep(s.wait)
	w.WriteHeader(res.StatusCode)
	if _, err := w.Write(body); err != nil {
		s.log.WithError(err).Error("Error writing response body")
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

	if r.ContentLength > 0 {
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

func (s *service) writeResponse(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response{
		Message: msg,
	}); err != nil {
		s.log.WithError(err).Errorf("Failed to write response")
	}
}

func (s *service) forwardRequest(r *http.Request, target string) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	return s.client.Do(req)
}
