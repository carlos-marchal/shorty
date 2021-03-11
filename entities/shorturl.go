package entities

import (
	"fmt"
	"net/url"
	"regexp"
	"time"
)

type ShortURL struct {
	Target  string
	ShortID string
	Expires time.Time
}

var idRegexp = regexp.MustCompile(`^[[:alnum:]]+$`)

type ErrInvalidURL struct {
	url string
}

func (err *ErrInvalidURL) Error() string {
	return fmt.Sprintf("URL %v is either not valid or not http(s)", err.url)
}

type ErrInvalidID struct {
	id string
}

func (err *ErrInvalidID) Error() string {
	return fmt.Sprintf("id %v is not alphanumeric", err.id)
}

func NewShortURL(target string, shortID string) (*ShortURL, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if parsedTarget.Scheme != "http" && parsedTarget.Scheme != "https" {
		return nil, &ErrInvalidURL{target}
	}
	if !idRegexp.MatchString(shortID) {
		return nil, &ErrInvalidID{shortID}
	}
	return &ShortURL{
		Target:  target,
		ShortID: shortID,
		Expires: time.Now().Add(time.Hour * 24 * 7),
	}, nil
}
