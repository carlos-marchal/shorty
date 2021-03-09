# Shorty

![Build status](https://github.com/carlos-marchal/shorty/actions/workflows/main.yaml/badge.svg)

Shorty is a proof of concept on using git backed storage, as well as an
implementation of [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) 
in Go. It is a very simple URL shortener, that backs its data to a git
repository instead of a traditional database.

It is fully functional, however there is still room for improvement, specially
regarding response times, which are currently limited by git latency.

## Usage

An instance of this program is deployed at https://shorty.carlos.marchal.page,
and connected to this very repo, to the file called urls.json. It doesn't make 
for very 'short' URLs, but as a testing endpoint it gets the job done.

To shorten an url, you have to POST a JSON object with a single `"url"` string
field to the endpoint `/shorten`. The response is a JSON object with the
target URL, the shortened URL and the expiration date. Currently it only
supports HTTP(S) URLs.

```bash
curl https://shorty.carlos.marchal.page/shorten \
  --data '{"url": "https://your.url.goes.here"}' \
  --header "content-type: application/json" \
  --request POST
```

After this, your URL should appear in the corresponding file in the repo. You
can retrieve with a GET to the shortened URL returned after creation. The
response will be a temporary redirect, which in browsers should lead you
seamlessly to the target page.

```bash
curl [the url you got from previous response] -i
```

## Testing, building and running

To run all the test suites using Docker Compose, run the following command in
the repository root:

```
docker-compose -f docker-compose.test.yaml up --build --exit-code-from test
```

To build the server binary you can use:

```
go build
```

or you can build it directly as a docker container using:

```
docker build -t shorty .
```

When running the server, it must be configured with environment variables:

| Name                | Required | Default                        | Description                                                    |
| ------------------- | -------- | ------------------------------ | -------------------------------------------------------------- |
| REPO_URL            | yes      |                                | An ssh URL to a git repo used to store the data                |
| REPO_PRIVATE_KEY    | yes      |                                | A PEM encoded private key with permission to push to the repo  |
| URL_FILE_PATH       | no       | urls.json                      | The file where to store the URLs in the repo                   |
| COMMIT_NAME         | no       | Shorty Bot                     | The commit author name of the bot                              |
| COMMIT_EMAIL        | no       | shorty.bot@carlos.marchal.page | The commit author email of the bot                             |
| PORT                | no       | 8080                           | The port on which to listen                                    |
| HOSTNAME            | no       | localhost                      | The hostname to use in responses                               |

