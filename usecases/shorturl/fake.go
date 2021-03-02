package shorturl

import (
	"fmt"

	"github.com/carlos-marchal/shorty/entities"
)

type fakeRepository struct {
	byID  map[string]*entities.ShortURL
	byURL map[string]*entities.ShortURL
	n     uint
}

func newfakeRepository() *fakeRepository {
	return &fakeRepository{
		byID:  make(map[string]*entities.ShortURL),
		byURL: make(map[string]*entities.ShortURL),
		n:     0,
	}
}

func (repository *fakeRepository) GetByURL(target string) (*entities.ShortURL, error) {
	url := repository.byURL[target]
	if url == nil {
		return nil, fmt.Errorf("no url with target %v in repo", url)
	}
	return url, nil
}

func (repository *fakeRepository) GetByID(shortID string) (*entities.ShortURL, error) {
	url := repository.byID[shortID]
	if url == nil {
		return nil, fmt.Errorf("no url with ID %v in repo", url)
	}
	return url, nil
}

func (repository *fakeRepository) SaveURL(url *entities.ShortURL) error {
	repository.byID[url.ShortID] = url
	repository.byURL[url.Target] = url
	return nil
}

func (repository *fakeRepository) GenerateShortID() (string, error) {
	repository.n++
	return fmt.Sprintf("%x", repository.n), nil
}

var _ Repository = (*fakeRepository)(nil)
