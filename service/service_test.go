package service_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/santihernandezc/devnull/service"
)

func TestForwardRequest(t *testing.T) {
	targetSrv := httptest.NewServer(http.HandlerFunc(handler))
	defer targetSrv.Close()

	tests := []struct {
		name          string
		target        string
		statusCode    int
		wait          time.Duration
		expStatusCode int
	}{
		{
			name:          "no target, custom status code",
			statusCode:    http.StatusNotAcceptable,
			expStatusCode: http.StatusNotAcceptable,
		},
	}

	l := log.New(io.Discard, "", 0)

	for _, test := range tests {
		svc := service.New(l, test.target, false, test.statusCode, test.wait)
		s := httptest.NewServer(http.HandlerFunc(svc.Handler))
		defer s.Close()

		res, err := http.Get(s.URL)
		if err != nil {
			t.Fatalf("Error making GET request: %v", err)
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Fatalf("Error closing request body: %v", err)
			}
		}()

		if res.StatusCode != test.expStatusCode {
			t.Fatalf("Unexpected status code, got=%d, want=%d", res.StatusCode, test.expStatusCode)
		}

		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading respone body: %v", err)
		}

		fmt.Println(string(b))
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}
