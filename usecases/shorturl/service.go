package shorturl

import (
	"time"

	"github.com/carlos-marchal/shorty/entities"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) (*Service, error) {
	return &Service{repository}, nil
}

func (service *Service) ShortenURL(target string) (*entities.ShortURL, error) {
	url, err := service.repository.GetByURL(target)
	switch err.(type) {
	case nil:
		return url, err
	case *ErrRepoNotFound:
		break
	default:
		return nil, err
	}
	id, err := service.repository.GenerateShortID()
	if err != nil {
		return nil, err
	}
	new, err := entities.NewShortURL(target, id)
	if err != nil {
		return nil, err
	}
	err = service.repository.SaveURL(new)
	if err != nil {
		return nil, err
	}
	return new, nil
}

func (service *Service) ResolveURL(shortID string) (*entities.ShortURL, error) {
	url, err := service.repository.GetByID(shortID)
	if err != nil {
		return nil, err
	}
	if url.Expires.Before(time.Now()) {
		return nil, &ErrURLExpired{url.Target, url.Expires}
	}
	return url, nil
}
