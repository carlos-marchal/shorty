package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/carlos-marchal/shorty/entities"
)

type fakeUserService struct {
	CallsToShorten []string
	CallsToResolve []string
}

func (service *fakeUserService) ShortenURL(target string) (*entities.ShortURL, error) {
	service.CallsToShorten = append(service.CallsToShorten, target)
	return entities.NewShortURL(target, "fakeid")
}

func (service *fakeUserService) ResolveURL(shortID string) (*entities.ShortURL, error) {
	service.CallsToResolve = append(service.CallsToResolve, shortID)
	return entities.NewShortURL("https://fake.com", shortID)
}

var testHandler = buildHandler(&fakeUserService{}, &Config{Port: 8000, Origin: "https://test"})

func TestShortenAcceptsOnlyPOST(t *testing.T) {
	tests := []struct {
		method   string
		expectOK bool
	}{
		{method: "GET", expectOK: false},
		{method: "POST", expectOK: true},
		{method: "PUT", expectOK: false},
		{method: "PATCH", expectOK: false},
		{method: "DELETE", expectOK: false},
		{method: "OPTIONS", expectOK: false},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, "/shorten", strings.NewReader(`{"url": "https://example.com"}`))
		request.Header.Set("content-type", "application/json")
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, request)
		var ok bool
		switch status := w.Result().StatusCode; status {
		case http.StatusOK:
			ok = true
		case http.StatusMethodNotAllowed:
			ok = false
		default:
			t.Fatalf("Unexpected status code %v", status)
		}
		if ok != test.expectOK {
			t.Fatalf("Expected ok to be %v but is %v", test.expectOK, ok)
		}
	}
}
func TestShortenRequestHasProperFormat(t *testing.T) {
	tests := []struct {
		contentType string
		content     string
		expectOK    bool
	}{
		{contentType: "text/plain", content: "hello!", expectOK: false},
		{contentType: "text/html", content: "<div>Hello!</div>", expectOK: false},
		{contentType: "application/json", content: "{ bad json ]", expectOK: false},
		{contentType: "application/json", content: `{"unexpected-field": "baad"}`, expectOK: false},
		{contentType: "application/json", content: `{"url": "https://example.com"}`, expectOK: true},
	}
	for _, test := range tests {
		request := httptest.NewRequest("POST", "/shorten", strings.NewReader(test.content))
		request.Header.Set("content-type", test.contentType)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, request)
		var ok bool
		switch status := w.Result().StatusCode; status {
		case http.StatusOK:
			ok = true
		case http.StatusBadRequest:
			ok = false
		default:
			t.Fatalf("Unexpected status code %v for case %+v", status, test)
		}
		if ok != test.expectOK {
			t.Fatalf("Expected ok to be %v but is %v", test.expectOK, ok)
		}
	}
}

func TestShortenResponseHasProperFormat(t *testing.T) {

}
