package shorturl

import (
	"fmt"
	"time"

	"github.com/carlos-marchal/shorty/entities"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("must pass a non nil repository")
	}
	return &Service{repository}, nil
}

func (service *Service) ShortenURL(target string) (*entities.ShortURL, error) {
	_, err := service.repository.GetByURL(target)
	if err == nil {
		return nil, fmt.Errorf("url %v has already been shortened", target)
	}
	id, err := service.repository.GenerateShortID()
	if err != nil {
		return nil, fmt.Errorf("error generating id: %v", err)
	}
	new, err := entities.NewShortURL(target, id)
	if err != nil {
		return nil, fmt.Errorf("error constructing URL: %v", err)
	}
	err = service.repository.SaveURL(new)
	if err != nil {
		return nil, fmt.Errorf("error saving URL: %v", err)
	}
	return new, nil
}

func (service *Service) ResolveURL(shortID string) (*entities.ShortURL, error) {
	url, err := service.repository.GetByID(shortID)
	if err != nil {
		return nil, err
	}
	if url.Expires.Before(time.Now()) {
		return nil, fmt.Errorf("entry for %v is expired", shortID)
	}
	return url, nil
}
