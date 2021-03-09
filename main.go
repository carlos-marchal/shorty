package main

import (
	"log"
	"os"
	"strconv"

	"github.com/carlos-marchal/shorty/git"
	"github.com/carlos-marchal/shorty/http"
	"github.com/carlos-marchal/shorty/usecases/shorturl"
)

var defaultEnv = map[string]string{
	"REPO_URL":         "",
	"REPO_PRIVATE_KEY": "",
	"URL_FILE_PATH":    "urls.json",
	"COMMIT_NAME":      "Shorty Bot",
	"COMMIT_EMAIL":     "shorty.bot@carlos.marchal.page",
	"PORT":             "8080",
	"ORIGIN":           "http://localhost:8080",
}

func main() {
	env := make(map[string]string)
	for key, defaultValue := range defaultEnv {
		var value string
		if passedValue := os.Getenv(key); passedValue != "" {
			value = passedValue
		} else if defaultValue != "" {
			value = defaultValue
		} else {
			log.Fatalf("You need to provide an env value for %v\n", key)
		}
		env[key] = value
	}
	repository, err := git.NewRepository(&git.Config{
		RepoURL:     env["REPO_URL"],
		PrivateKey:  env["REPO_PRIVATE_KEY"],
		URLFilePath: env["URL_FILE_PATH"],
		CommitName:  env["COMMIT_NAME"],
		CommitEmail: env["COMMIT_EMAIL"],
	})
	if err != nil {
		log.Fatalf("Error initializing repository: %v", err)
	}
	service, err := shorturl.NewService(repository)
	if err != nil {
		log.Fatalf("Error initializing use case handler: %v", err)
	}
	port, err := strconv.ParseUint(env["PORT"], 10, 16)
	if err != nil {
		log.Fatalf("Error parsing port number: %v", env["PORT"])
	}
	err = http.Start(service, &http.Config{
		Port:   uint(port),
		Origin: env["ORIGIN"],
	})
	log.Fatalf("Error initializing use case handler: %v", err)
}
