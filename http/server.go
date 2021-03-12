package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/carlos-marchal/shorty/entities"
	"github.com/carlos-marchal/shorty/usecases/shorturl"
)

type requestBody struct {
	URL *string `json:"url"`
}

type responseBody struct {
	Target    string    `json:"target"`
	Shortened string    `json:"shortened"`
	Expires   time.Time `json:"expires"`
}

type Config struct {
	Port   uint
	Origin string
}

func Start(urls shorturl.UseCase, config *Config) error {
	handler := buildHandler(urls, config)
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", config.Port), handler)
}

func sendErrorJSON(w http.ResponseWriter, error string, code int) {
	body, _ := json.MarshalIndent(struct {
		Error string `json:"error"`
	}{error}, "", "  ")
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(body))
}

func buildHandler(urls shorturl.UseCase, config *Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			sendErrorJSON(w, fmt.Sprintf("method %v not supported", r.Method), http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("content-type") != "application/json" {
			sendErrorJSON(w, "body must be a json object with a single url string field", http.StatusBadRequest)
			return
		}
		parsed := new(requestBody)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(parsed)
		if err != nil || parsed.URL == nil || decoder.More() {
			sendErrorJSON(w, "body must be a json object with a single url string field", http.StatusBadRequest)
			return
		}
		log.Printf("%+v\n", parsed)
		url, err := urls.ShortenURL(*parsed.URL)
		switch err.(type) {
		case *entities.ErrInvalidURL:
			sendErrorJSON(w, "URL must be a valid HTTP or HTTPS URL", http.StatusBadRequest)
			return
		case nil:
			break
		default:
			sendErrorJSON(w, "internal server error", http.StatusInternalServerError)
			return
		}
		response := &responseBody{
			Target:    url.Target,
			Shortened: buildURL(config.Origin, config.Port, url.ShortID),
			Expires:   url.Expires,
		}
		responseBody, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			sendErrorJSON(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Add("content-type", "application/json")
		w.Write(responseBody)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			sendErrorJSON(w, fmt.Sprintf("method %v not supported", r.Method), http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Path[1:]
		if id == "" {
			sendErrorJSON(w, "must provide an id to resolve", http.StatusBadRequest)
			return
		}
		url, err := urls.ResolveURL(id)
		if err != nil {
			sendErrorJSON(w, "url not found", http.StatusNotFound)
			return
		}
		shortened := buildURL(config.Origin, config.Port, url.ShortID)
		http.Redirect(w, r, shortened, http.StatusTemporaryRedirect)
	})

	return mux
}

func buildURL(origin string, port uint, path string) string {
	canonicalHTTP := origin[0:7] == "http://" && port == 80
	canonicalHTTPS := origin[0:8] == "https://" && port == 443
	if canonicalHTTP || canonicalHTTPS {
		return fmt.Sprintf("%v/%v", origin, path)
	}
	return fmt.Sprintf("%v:%v/%v", origin, port, path)
}
