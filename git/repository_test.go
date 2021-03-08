// +build integration

package git

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/carlos-marchal/shorty/entities"
)

var emptyRepoConfig *Config
var exampleRepoConfig *Config

func TestMain(m *testing.M) {
	key, err := ioutil.ReadFile("./test/ssh/client/client")
	if err != nil {
		os.Exit(-1)
	}
	emptyRepoConfig = &Config{
		PrivateKey: string(key),
		RepoURL:    "git@gitserver:/home/git/empty.git",
		// RepoURL:     "ssh://git@localhost:2222/home/git/empty.git",
		URLFilePath: "urls.json",
		CommitName:  "Shorty Bot Test",
		CommitEmail: "test@example.com",
	}
	exampleRepoConfig = new(Config)
	*exampleRepoConfig = *emptyRepoConfig
	exampleRepoConfig.RepoURL = "git@gitserver:/home/git/example.git"
	// exampleRepoConfig.RepoURL = "ssh://git@localhost:2222/home/git/example.git"
	os.Exit(m.Run())
}

func TestGetsFromRepoWithNoURLFile(t *testing.T) {
	_, err := NewRepository(emptyRepoConfig)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAssignsDistinctIDs(t *testing.T) {
	repo, err := NewRepository(emptyRepoConfig)
	if err != nil {
		t.Fatal(err)
	}
	id1, err := repo.GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	id2, err := repo.GenerateShortID()
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatalf("did not generate distinct IDs, got ID %v", id1)
	}
}

func TestStoresAndRetreivesCorrectly(t *testing.T) {
	repo, err := NewRepository(emptyRepoConfig)
	if err != nil {
		t.Fatal(err)
	}
	target, id := "https://www.example.com", "shortid"
	url, err := entities.NewShortURL(target, id)
	if err != nil {
		t.Fatal(err)
	}
	err = repo.SaveURL(url)
	if err != nil {
		t.Fatal(err)
	}
	byID, err := repo.GetByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if *url != *byID {
		t.Fatalf("expected: %+v, got: %+v", url, byID)
	}
	byURL, err := repo.GetByURL(target)
	if err != nil {
		t.Fatal(err)
	}
	if *url != *byURL {
		t.Fatalf("expected: %+v, got: %+v", url, byURL)
	}
}

func TestReadsExistentRepoCorrectly(t *testing.T) {
	repo, err := NewRepository(exampleRepoConfig)
	if err != nil {
		t.Fatal(err)
	}
	url, err := repo.GetByID("googleid")
	if err != nil {
		t.Fatal(err)
	}
	if url == nil {
		t.Fatal("got nil url from existing repo")
	}
}

func TestPreservesURLsWhenAddingToNonEmptyRepo(t *testing.T) {
	repo, err := NewRepository(exampleRepoConfig)
	if err != nil {
		t.Fatal(err)
	}
	newURL, err := entities.NewShortURL("https://wikipedia.org", "wikiid")
	if err != nil {
		t.Fatal(err)
	}
	err = repo.SaveURL(newURL)
	url, err := repo.GetByID("googleid")
	if err != nil {
		t.Fatal(err)
	}
	if url == nil {
		t.Fatal("got nil url from existing repo")
	}
	url, err = repo.GetByID("wikiid")
	if err != nil {
		t.Fatal(err)
	}
	if url == nil {
		t.Fatal("got nil url from existing repo")
	}
}
