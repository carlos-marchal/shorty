package entities

import (
	"testing"
	"time"
)

type testData struct {
	target  string
	shortID string
	errors  bool
}

func TestAssignsValues(t *testing.T) {
	target, shortID := "https://example.com", "id"
	url, err := NewShortURL(target, shortID)
	if err != nil {
		t.Fatalf("unexpected error %+v", err)
	}
	if url.Target != target || url.ShortID != shortID {
		t.Fatalf("failed sanity test, values not assigned correctly: %+v", url)
	}
}

func TestRejectsInvalidURL(t *testing.T) {
	_, err := NewShortURL("!😭invalid-url", "id")
	if err == nil {
		t.Fatalf("accepted invalid url")
	}
}

func TestRejectsNonHttpURL(t *testing.T) {
	urls := []string{
		"ftp://username:password@ftp.example.com",
		"file:///home/user/",
		"data:image/gif;base64,R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw== ",
	}
	for _, url := range urls {
		_, err := NewShortURL(url, "abc")
		if err == nil {
			t.Fatalf("expected error for non http(s) url %v", url)
		}
	}
}

func TestAcceptsOnlyAlphanumericIDs(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{id: "", valid: false},
		{id: "a", valid: true},
		{id: "Z", valid: true},
		{id: "1", valid: true},
		{id: "abcABC123", valid: true},
		{id: "abc-123", valid: false},
		{id: "🌵🌳🌲", valid: false},
	}
	for _, test := range tests {
		_, err := NewShortURL("https://example.com", test.id)
		if err == nil && !test.valid {
			t.Fatalf("accepted id %v when supposed to error", test.id)
		} else if err != nil && test.valid {
			t.Fatalf("threw error %v for id %v when supposed to accept", err, test.id)

		}
	}
}

func TestShortenedURLLastsOneWeek(t *testing.T) {
	url, err := NewShortURL("https://example.com", "abc")
	if err != nil {
		t.Fatalf("encountered error %v", err)
	}
	nextWeek := time.Now().Add(time.Hour * 24 * 7)
	diff := url.Expires.Sub(nextWeek)
	if diff < -time.Second || diff > time.Second {
		t.Fatalf("incorrect expiration date set at %v", url.Expires)
	}
}
