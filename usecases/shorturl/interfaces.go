package shorturl

import (
	"fmt"
	"time"

	"github.com/carlos-marchal/shorty/entities"
)

type Repository interface {
	GetByURL(shortID string) (*entities.ShortURL, error)
	GetByID(shortID string) (*entities.ShortURL, error)
	GenerateShortID() (string, error)
	SaveURL(url *entities.ShortURL) error
}

type ErrRepoNotFound struct {
	ID string
}

func (err *ErrRepoNotFound) Error() string {
	return fmt.Sprintf("identifier %v not found in repo", err.ID)
}

type ErrRepoInternal struct{}

func (err *ErrRepoInternal) Error() string {
	return "internal repo error"
}

type UseCase interface {
	ShortenURL(target string) (*entities.ShortURL, error)
	ResolveURL(shortID string) (*entities.ShortURL, error)
}

type ErrURLExpired struct {
	URL  string
	Time time.Time
}

func (err *ErrURLExpired) Error() string {
	return fmt.Sprintf("url %v expired on %v", err.URL, err.Time)
}
