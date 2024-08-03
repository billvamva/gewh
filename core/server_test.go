package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	t.Run("test sending data to server", func(t *testing.T) {
		svr := MessageServer()
		
		body := strings.NewReader("hello, world")

		request := httptest.NewRequest(http.MethodPost, "/", body)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)
	})
}