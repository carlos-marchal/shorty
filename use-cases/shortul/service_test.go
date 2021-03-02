package shorturl

import (
	"testing"
	"time"

	"github.com/carlos-marchal/shorty/entities"
)

func TestStoresAndRetrievesURLs(t *testing.T) {
	service := Service{newfakeRepository()}
	stored, err := service.ShortenURL("https://example.com")
	if err != nil {
		t.Fatalf("did not expect error while storing: %v", err)
	}
	retrieved, err := service.ResolveURL(stored.ShortID)
	if err != nil {
		t.Fatalf("did not expect error while retrieving: %v", err)
	}
	if stored.Target != retrieved.Target {
		t.Fatalf("expected %v to equal %v", stored.Target, retrieved.Target)
	}
}

func TestIgnoresExpiredURLs(t *testing.T) {
	service := Service{newfakeRepository()}
	service.repository.SaveURL(&entities.ShortURL{
		Target:  "https://example.com",
		ShortID: "id",
		Expires: time.Now().Add(-time.Second),
	})
	retrieved, err := service.ResolveURL("id")
	if err == nil {
		t.Fatalf("expected error on retrieving expired url, got %v", retrieved)
	}
}

func TestErrorsOnRepeatedEntry(t *testing.T) {
	service := Service{newfakeRepository()}
	first, err := service.ShortenURL("https://example.com")
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	second, err := service.ShortenURL("https://example.com")
	if err == nil {
		t.Fatalf("expected error on repeated entries, got entries %v and %v", first, second)
	}
}
