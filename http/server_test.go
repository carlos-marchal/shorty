package http

import (
	"github.com/carlos-marchal/shorty/entities"
)

type fakeUserService struct {
	CallsToShorten []string
	CallsToResolve []string
}

func (service *fakeUserService) ShortenURL(target string) (*entities.ShortURL, error) {
	service.CallsToShorten = append(service.CallsToShorten, target)
	return entities.NewShortURL(target, "fakeid")
}

func (service *fakeUserService) ResolveURL(shortID string) (*entities.ShortURL, error) {
	service.CallsToResolve = append(service.CallsToResolve, shortID)
	return entities.NewShortURL("https://fake.com", shortID)
}
