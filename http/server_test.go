package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/carlos-marchal/shorty/entities"
)

type fakeUserService struct {
	resultURL   *entities.ShortURL
	resultError error
	custom      bool
}

var defaultTestResponse = &entities.ShortURL{"http://example.com", "1", time.Now()}

func (service *fakeUserService) ShortenURL(target string) (*entities.ShortURL, error) {
	if service.custom {
		return service.resultURL, service.resultError
	}
	return defaultTestResponse, nil
}

func (service *fakeUserService) ResolveURL(shortID string) (*entities.ShortURL, error) {
	if service.custom {
		return service.resultURL, service.resultError
	}
	return defaultTestResponse, nil
}

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
		testHandler := buildHandler(&fakeUserService{}, &Config{Origin: "https://test"})
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
		fakeUserService
	}{
		{contentType: "text/plain", content: "hello!", expectOK: false},
		{contentType: "text/html", content: "<div>Hello!</div>", expectOK: false},
		{contentType: "application/json", content: "{ bad json ]", expectOK: false},
		{contentType: "application/json", content: `{"unexpected-field": "baad"}`, expectOK: false},
		{contentType: "application/json", content: `{"url": "https://example.com"}`, expectOK: true},
		{contentType: "application/json", content: `{"url": "ftp://example.com"}`, expectOK: false,
			fakeUserService: fakeUserService{
				custom:      true,
				resultError: &entities.ErrInvalidURL{},
			}},
		{contentType: "application/json", content: `{"url": "https://example.com", "unexpected-field": "baad"}`, expectOK: false},
		{contentType: "text/plain", content: `{"url": "https://example.com"}`, expectOK: false},
	}
	for _, test := range tests {
		request := httptest.NewRequest("POST", "/shorten", strings.NewReader(test.content))
		request.Header.Set("content-type", test.contentType)
		w := httptest.NewRecorder()
		testHandler := buildHandler(&test, &Config{Origin: "https://test"})
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
			t.Fatalf("Expected ok to be %v but is %v for %+v", test.expectOK, ok, test)
		}
	}
}

func TestShortenResponseHasProperFormat(t *testing.T) {
	request := httptest.NewRequest("POST", "/shorten", strings.NewReader(`{"url": "https://example.com"}`))
	request.Header.Set("content-type", "application/json")
	w := httptest.NewRecorder()
	testHandler := buildHandler(&fakeUserService{}, &Config{Origin: "https://test"})
	testHandler.ServeHTTP(w, request)
	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected ok status but got %v", response.StatusCode)
	}
	if mime := response.Header.Get("content-type"); mime != "application/json" {
		t.Fatalf("Expected json content type but got %v", mime)
	}
	parsed := new(responseBody)
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(parsed)
	if err != nil || decoder.More() {
		t.Fatalf("Expected response to match json schema")
	}
}

func TestResolveAcceptsOnlyGet(t *testing.T) {
	tests := []struct {
		method   string
		expectOK bool
	}{
		{method: "GET", expectOK: true},
		{method: "POST", expectOK: false},
		{method: "PUT", expectOK: false},
		{method: "PATCH", expectOK: false},
		{method: "DELETE", expectOK: false},
		{method: "OPTIONS", expectOK: false},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, "/id", nil)
		w := httptest.NewRecorder()
		testHandler := buildHandler(&fakeUserService{}, &Config{Origin: "https://test"})
		testHandler.ServeHTTP(w, request)
		var ok bool
		switch status := w.Result().StatusCode; status {
		case http.StatusTemporaryRedirect:
			ok = true
		case http.StatusMethodNotAllowed:
			ok = false
		default:
			t.Fatalf("Unexpected status code %v", status)
		}
		if ok != test.expectOK {
			t.Fatalf("Expected ok to be %v but is %v for method %v", test.expectOK, ok, test.method)
		}
	}
}

func TestResolveRedirectsToURL(t *testing.T) {
	tests := []struct {
		origin  string
		shortID string
		target  string
	}{
		{origin: "http://example.com", shortID: "id", target: "http://other.example.com"},
		{origin: "https://example.com", shortID: "12345", target: "http://other.example.com"},
		{origin: "https://example.com:1247", shortID: "alph1num4ric", target: "http://other.example.com"},
		{origin: "http://example.com:2", shortID: "qw-er-ty", target: "http://other.example.com"},
	}
	for _, test := range tests {
		request := httptest.NewRequest("GET", fmt.Sprintf("/%v", test.shortID), nil)
		w := httptest.NewRecorder()
		testHandler := buildHandler(&fakeUserService{
			custom:    true,
			resultURL: &entities.ShortURL{ShortID: test.shortID, Target: test.target},
		},
			&Config{Origin: test.origin},
		)

		testHandler.ServeHTTP(w, request)
		if status := w.Result().StatusCode; status != http.StatusTemporaryRedirect {
			t.Fatalf("Unexpected status code %v", status)
		}
		actual := w.Header().Get("location")
		if test.target != actual {
			t.Fatalf("Expected %v but got %v", test.target, actual)
		}
	}
}
