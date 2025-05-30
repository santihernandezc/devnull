package service_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/santihernandezc/devnull/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	targetServerHeaders := map[string][]string{
		"Target-Header-1": []string{"test-1"},
		"Target-Header-2": []string{"test-2"},
	}
	targetServerResponse := []byte("this is a test response")
	ft := newFakeTarget(targetServerResponse, targetServerHeaders)
	targetSrv := httptest.NewServer(http.HandlerFunc(ft.handler))
	defer targetSrv.Close()

	tests := []struct {
		name          string
		headers       map[string][]string
		statusCode    int
		target        string
		wait          time.Duration
		expStatusCode int
		expHeaders    map[string][]string
		expResponse   []byte
	}{
		{
			name:          "no target, default status code",
			expStatusCode: http.StatusOK,
		},
		{
			name:          "no target, custom status code",
			statusCode:    http.StatusNotAcceptable,
			expStatusCode: http.StatusNotAcceptable,
		},
		{
			name:          "no target, custom wait time",
			expStatusCode: http.StatusOK,
			wait:          time.Second,
		},
		{
			name:          "no target, bad status code",
			statusCode:    19,
			expStatusCode: http.StatusOK,
			wait:          time.Second,
		},
		{
			name:          "bad target, custom wait time",
			target:        "http:bad",
			expStatusCode: http.StatusBadGateway,
			wait:          time.Second,
		},
		{
			name:          "target configured, custom wait time",
			target:        targetSrv.URL,
			expStatusCode: http.StatusTeapot,
			wait:          time.Second,
			expHeaders:    targetServerHeaders,
			expResponse:   targetServerResponse,
		},
		{
			name:          "target configured, custom status code ignored",
			target:        targetSrv.URL,
			statusCode:    http.StatusAccepted,
			expStatusCode: http.StatusTeapot,
			expHeaders:    targetServerHeaders,
			expResponse:   targetServerResponse,
		},
		{
			name:          "target configured, custom headers",
			target:        targetSrv.URL,
			headers:       map[string][]string{"Test-Header-1": []string{"test-1"}, "Test-Header-2": []string{"test-2"}},
			expStatusCode: http.StatusTeapot,
			expHeaders:    targetServerHeaders,
			expResponse:   targetServerResponse,
		},
	}

	l := logrus.New()
	l.SetOutput(io.Discard)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(ft.cleanup)
			svc := service.New(l, test.target, false, test.statusCode, test.wait)
			s := httptest.NewServer(http.HandlerFunc(svc.Handler))
			defer s.Close()

			now := time.Now()
			req, err := http.NewRequest(http.MethodGet, s.URL+"/test", nil)
			require.NoError(t, err)

			for k, v := range test.headers {
				for _, vv := range v {
					req.Header.Add(k, vv)
				}
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, test.expStatusCode, res.StatusCode)

			elapsed := time.Since(now)
			require.Greater(t, elapsed, test.wait)

			for k, v := range test.headers {
				header := ft.receivedHeaders[k]
				for _, vv := range v {
					require.Contains(t, header, vv)
				}
			}

			for k, v := range test.expHeaders {
				header := res.Header[k]
				for _, vv := range v {
					require.Contains(t, header, vv)
				}
			}

			if len(test.expResponse) > 0 {
				defer func() {
					require.NoError(t, res.Body.Close())
				}()

				b, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, test.expResponse, b)
			}
		})
	}
}

type fakeTarget struct {
	receivedHeaders map[string][]string
	response        []byte
	responseHeaders map[string][]string
}

func newFakeTarget(response []byte, headers map[string][]string) *fakeTarget {
	return &fakeTarget{
		receivedHeaders: make(map[string][]string),
		response:        response,
		responseHeaders: headers,
	}
}

func (f *fakeTarget) handler(w http.ResponseWriter, r *http.Request) {
	for k, v := range f.responseHeaders {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(http.StatusTeapot)
	f.receivedHeaders = r.Header
	w.Write(f.response)
}

func (f *fakeTarget) cleanup() {
	f.receivedHeaders = make(map[string][]string)
}
