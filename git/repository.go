package git

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/carlos-marchal/shorty/entities"
	"github.com/carlos-marchal/shorty/usecases/shorturl"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Config struct {
	RepoURL     string
	PrivateKey  string
	URLFilePath string
	CommitName  string
	CommitEmail string
}

type Repository struct {
	config      *Config
	repository  *git.Repository
	worktree    *git.Worktree
	fs          billy.Filesystem
	urls        []*entities.ShortURL
	urlByID     map[string]*entities.ShortURL
	urlByTarget map[string]*entities.ShortURL
	serial      uint
	keys        *ssh.PublicKeys
}

type urlFileType struct {
	URLs   []*entities.ShortURL
	Serial uint
}

func (repository *Repository) readRemote() error {
	err := repository.repository.Fetch(&git.FetchOptions{
		Auth:       repository.keys,
		RemoteName: "origin",
		Depth:      1,
	})
	switch err {
	case nil:
		err = repository.worktree.Pull(&git.PullOptions{
			Auth:       repository.keys,
			RemoteName: "origin",
			Depth:      1,
		})
		if err != nil {
			return fmt.Errorf("error pulling remote: %w", err)
		}
		return repository.readRemoteNoFetch()
	case git.NoErrAlreadyUpToDate:
		return nil
	case transport.ErrEmptyRemoteRepository:
		return nil
	default:
		return fmt.Errorf("error fetching remote: %w", err)
	}
}

func (repository *Repository) readRemoteNoFetch() error {
	urlFileContent, err := repository.fs.Open(repository.config.URLFilePath)
	if err != nil {
		if err.Error() == "file does not exist" {
			repository.urls = []*entities.ShortURL{}
			repository.serial = 0
		} else {
			return fmt.Errorf("error opening url file: %w", err)
		}
	} else {
		rawContent, err := io.ReadAll(urlFileContent)
		if err != nil {
			return fmt.Errorf("error reading url file: %w", err)
		}
		err = urlFileContent.Close()
		if err != nil {
			return fmt.Errorf("error closing url file: %w", err)
		}
		urlFile := new(urlFileType)
		err = json.Unmarshal(rawContent, urlFile)
		if err != nil {
			return fmt.Errorf("error deserializing url file: %w", err)
		}
		repository.urls = urlFile.URLs
		repository.serial = urlFile.Serial
	}
	repository.urlByID = make(map[string]*entities.ShortURL)
	repository.urlByTarget = make(map[string]*entities.ShortURL)
	for _, url := range repository.urls {
		repository.urlByID[url.ShortID] = url
		repository.urlByTarget[url.Target] = url
	}
	return nil
}

func (repository *Repository) writeRemote(commitMessage string) error {
	file, err := repository.fs.OpenFile(repository.config.URLFilePath, os.O_RDWR, 666)
	if err != nil {
		if err.Error() == "file does not exist" {
			file, err = repository.fs.Create(repository.config.URLFilePath)
			if err != nil {
				return fmt.Errorf("error creating url file: %w", err)
			}
		} else {
			return fmt.Errorf("error opening url file: %w", err)
		}
	}
	for i := 0; i < len(repository.urls); i++ {
		url := repository.urls[i]
		if url.Expires.Before(time.Now()) {
			repository.urls = append(repository.urls[:i], repository.urls[i+1:]...)
			delete(repository.urlByID, url.ShortID)
			delete(repository.urlByTarget, url.Target)
		}
	}
	urlFile := &urlFileType{repository.urls, repository.serial}
	fileContents, err := json.MarshalIndent(urlFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing url file: %w", err)
	}
	_, err = file.Write(fileContents)
	if err != nil {
		return fmt.Errorf("error writing url file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}
	_, err = repository.worktree.Add(repository.config.URLFilePath)
	if err != nil {
		return fmt.Errorf("error adding file to repo: %w", err)
	}
	_, err = repository.worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  repository.config.CommitName,
			Email: repository.config.CommitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("error commiting file to repo: %w", err)
	}
	err = repository.repository.Push(&git.PushOptions{Auth: repository.keys, RemoteName: "origin"})
	if err != nil {
		return fmt.Errorf("error pushing file to repo: %w", err)
	}
	return nil
}

func NewRepository(config *Config) (*Repository, error) {
	keys, err := ssh.NewPublicKeys("git", []byte(config.PrivateKey), "")
	if err != nil {
		return nil, fmt.Errorf("error creating key: %w", err)
	}
	fs := memfs.New()
	storer := memory.NewStorage()
	gitRepo, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:   config.RepoURL,
		Auth:  keys,
		Depth: 1,
	})
	if err != nil && err != transport.ErrEmptyRemoteRepository {
		return nil, fmt.Errorf("error cloning repo: %w", err)
	}
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting worktree: %w", err)
	}
	repository := &Repository{
		config:      config,
		repository:  gitRepo,
		worktree:    worktree,
		fs:          fs,
		urls:        make([]*entities.ShortURL, 0),
		urlByID:     make(map[string]*entities.ShortURL),
		urlByTarget: make(map[string]*entities.ShortURL),
		serial:      0,
		keys:        keys,
	}
	err = repository.readRemoteNoFetch()
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func (repository *Repository) GetByURL(target string) (*entities.ShortURL, error) {
	err := repository.readRemote()
	if err != nil {
		return nil, err
	}
	url := repository.urlByTarget[target]
	if url == nil {
		return nil, fmt.Errorf("no url with target %v in repo", target)
	}
	return url, nil
}

func (repository *Repository) GetByID(shortID string) (*entities.ShortURL, error) {
	err := repository.readRemote()
	if err != nil {
		return nil, err
	}
	url := repository.urlByID[shortID]
	if url == nil {
		return nil, fmt.Errorf("no url with short id %v in repo", shortID)
	}
	return url, nil
}

func (repository *Repository) GenerateShortID() (string, error) {
	err := repository.readRemote()
	if err != nil {
		return "", err
	}
	id := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(fmt.Sprint(repository.serial)))
	repository.serial++
	err = repository.writeRemote(fmt.Sprintf("Increasing serial number to %v", repository.serial))
	if err != nil {
		repository.serial--
		return "", err
	}
	return id, nil
}

func (repository *Repository) SaveURL(url *entities.ShortURL) error {
	err := repository.readRemote()
	if err != nil {
		return err
	}
	repository.urls = append([]*entities.ShortURL{url}, repository.urls...)
	repository.urlByID[url.ShortID] = url
	repository.urlByTarget[url.Target] = url
	err = repository.writeRemote(fmt.Sprintf("Adding URL %v to list", url.Target))
	if err != nil {
		repository.urls = repository.urls[1:]
		delete(repository.urlByID, url.ShortID)
		delete(repository.urlByTarget, url.Target)
		return err
	}
	return nil
}

var _ shorturl.Repository = (*Repository)(nil)
