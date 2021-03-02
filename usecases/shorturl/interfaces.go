package shorturl

import "github.com/carlos-marchal/shorty/entities"

type Repository interface {
	GetByURL(shortID string) (*entities.ShortURL, error)
	GetByID(shortID string) (*entities.ShortURL, error)
	GenerateShortID() (string, error)
	SaveURL(url *entities.ShortURL) error
}

type UseCase interface {
	ShortenURL(target string) (*entities.ShortURL, error)
	ResolveURL(shortID string) (*entities.ShortURL, error)
}
