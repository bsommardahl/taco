# taco 🌮🌮🌮 [![CircleCI](https://circleci.com/gh/sul-dlss-labs/taco.svg?style=svg)](https://circleci.com/gh/sul-dlss-labs/taco)
The next generation repository system for DLSS
![taco](https://user-images.githubusercontent.com/92044/34897877-016a4e36-f7b6-11e7-80e3-4edecfb2f89d.gif)

## Swagger API

This configuration is for AWS API Gateway.  It was retrieved by going to the API, selecting the "prod" under "Stages" and then doing "Export" and selecting "Export as Swagger + API Gateway Extensions"

## Go Local Development Setup

1. Install go (grab binary from here or use `brew install go` on Mac OSX).
2. Setup your Go workspace (where your Go code, binaries, etc. are kept together. See some helpful documentation here on the Go concept of workspaces: https://github.com/alco/gostart#1-go-tool-is-only-compatible-with-code-that-resides-in-a-workspace.):
      ```bash
      $ mkdir -p ~/go
      $ export GOPATH=~/go
      $ export PATH=~/go/bin:$PATH
      $ cd ~/go
      ```
      Your Go code repositories will reside within `~/go/src/...` in the `$GOPATH`. Name these paths to avoid library clash, for example TACO Go code could be in `~/go/src/github.com/sul-dlss-labs/taco`. This should be where your Github repository resides too.
3. In order to download the project code to `~/go/src/github.com/sul-dlss-labs/taco`, from any directory in your ``$GOPATH`, run:
    ```bash
    $ go get github.com/sul-dlss-labs/taco
    ```
4. Handle Go project dependencies with the Go `dep` package:
    * Install Go Dep via `brew install dep` then `brew upgrade dep` (if on Mac OSX).
    * If your project's `Gopkg.toml` and `Gopkg.lock` have not yet been populated, you need to add an inferred list of your dependencies by running `dep init`.
    * If your project has those file populated, then make sure your dependencies are synced via running `dep ensure`.
    * If you need to add a new dependency, run `dep ensure -add github.com/pkg/errors`. This should add the dependency and put the new dependency in your `Gopkg.*` files.

5. Localstack and Environment Variables
    * Local development depends on [localstack](https://github.com/localstack/localstack) to mock the AWS environment.
    * Environment Variables for development default to [localstack](https://github.com/localstack/localstack), if these are run on non-standard ports, review the environment variables in [config](config/config.go)

## Running the Go Code locally without a build

```shell
$ cd cmd/tacod
$ AWS_ACCESS_KEY_ID=999999 AWS_SECRET_KEY=1231 go run main.go
```

## Building to TACO Binary

### Building for Docker
```shell
$ docker build -t taco  .
$ docker run -p 8080:8080 taco
```

### Build for the local OS
This will create a binary in your path that you can then run the application with.

```shell
$ go build -o tacod cmd/tacod/main.go
$ ./tacod
```

## Testing
```shell
$ go test -v ./...
```

## Running the TACO Binary
If you are running locally, we are stubbing out AWS services using the library `localstack`. See more information on installing `localstack` here: https://github.com/localstack/localstack#installing.

First start up DynamoDB locally via localstack:
```shell
$ SERVICES=dynamodb localstack start
```

Then create the table:
```shell
$ awslocal dynamodb create-table --table-name resources \
  --attribute-definitions "AttributeName=id,AttributeType=S" \
  --key-schema "AttributeName=id,KeyType=HASH" \
  --provisioned-throughput=ReadCapacityUnits=100,WriteCapacityUnits=100
```

Now start the API server:
```shell
% AWS_ACCESS_KEY_ID=999999 AWS_SECRET_KEY=1231 ./tacod
```

Then you can interact with it using `curl`:
```shell
curl -X POST -H "Content-Type: application/json" -d '{"title":"value1", "sourceId":"value2"}' http://localhost:8080/v1/resource
```

it will return a response like:
```json
{"id":"fe1f66a9-5285-4b28-8240-0482c8fff6c7"}
```

Then you can use the returned identifier to retrieve the original:

```shell
curl -H "Content-Type: application/json"  http://localhost:8080/v1/resource/fe1f66a9-5285-4b28-8240-0482c8fff6c7
```

## API Code Structure

We use `go-swagger` to generate the API code within the `generated/` directory.

We connect the specification-generated API code to our own handlers defined with `handlers/`. Handlers are where we add our own logic for processing requests.

Our handlers and the generated API code is connected within `main.go`, which is the file to start the API.

### Install Go-Swagger

The API code is generated from `swagger.yml` using `go-swagger` library. As this is not used in the existing codebase anywhere currently, you'll need to install the `go-swagger` library before running these commands (commands for those using Mac OSX):

```shell
brew tap go-swagger/go-swagger
brew install go-swagger
brew upgrade go-swagger
```

This should give you the `swagger` binary command in your $GOPATH and allow you to manage versions better (TBD write this up). The version of your go-swagger binary is **0.13.0** (run `swagger version` to check this).

### Nota Bene on go-swagger install from source

If instead of brew, you decide to install go-swagger from source (i.e. `go get -u github.com/go-swagger/go-swagger/cmd/swagger`), you will be running the library at its Github `dev` branch. You will need to change into that library (`cd $GOPATH/src/github.com/go-swagger/go-swagger`) and checkout from Github the latest release (`git checkout tags/0.13.0`). Then you will need to run `go install github.com/go-swagger/go-swagger/cmd/swagger` to generate the go-swagger binary in your `$GOPATH/bin`.

### To generate the API code

There appears to be no best way to handle specification-based re-generation of the `generated/` API code, so the following steps are recommended:

```shell
$ rm -rf generated/
$ swagger generate server -t generated --exclude-main
```

### Non-generated code

Anything outside of `generated/` is our own code and should not be touched by a regeneration of the API code from the Swagger specification.

## SWAGGER Generated Documentation

To see the SWAGGER generated documentation, go to https://sul-dlss-labs.github.io/taco/. This is automatically generated off of the Swagger spec in this repository on the `master` branch.

If you want to generate this documentation locally, you can run the following:

```shell
$ swagger serve swagger.yml
```

This should prompt you to your web browser for the HTML generated docs.
