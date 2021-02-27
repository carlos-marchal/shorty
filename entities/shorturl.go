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

func NewShortURL(target string, shortID string) (*ShortURL, error) {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if parsedTarget.Scheme != "http" && parsedTarget.Scheme != "https" {
		return nil, fmt.Errorf("scheme must be either http or https")
	}
	parsedShortened, err := url.Parse(shortID)
	if err != nil {
		return nil, err
	}
	if parsedShortened.IsAbs() {
		return nil, fmt.Errorf("shortenedpath must be a relative path")

	}
	if !idRegexp.MatchString(shortID) {
		return nil, fmt.Errorf("shortened id %v is not alphanumeric", shortID)
	}
	return &ShortURL{
		Target:  target,
		ShortID: shortID,
		Expires: time.Now().Add(time.Hour * 24 * 7),
	}, nil
}
