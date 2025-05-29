package service

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
			s.log.WithError(err).
				WithField("target", s.target).
				WithField("path", r.URL.Path).
				Error("Error joining paths")
			return
		}

		res, err := forwardRequest(r, uri)
		if err != nil {
			s.log.WithError(err).WithField("target", s.target).Error("Error forwarding request")
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
		return
	}

	time.Sleep(s.wait)
	if s.statusCode != 200 {
		w.WriteHeader(s.statusCode)
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

	switch r.Method {
	case http.MethodPatch, http.MethodPost, http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.log.WithError(err).Error("Error reading request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		s.log.WithFields(
			logrus.Fields{
				"method":  r.Method,
				"uri":     r.RequestURI,
				"headers": strings.Join(headers, ", "),
				"body":    string(body),
			}).Info("Request received")
	default:
		s.log.WithFields(logrus.Fields{
			"method":  r.Method,
			"uri":     r.RequestURI,
			"headers": strings.Join(headers, ", "),
		}).Info("Request received")
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
