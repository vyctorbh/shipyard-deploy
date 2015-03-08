CGO_ENABLED=0
GOOS=linux
GOARCH=amd64
TAG=${1:-latest}
REPO=shipyard/deploy

all: build

deps:
	@godep restore

build: deps
	@godep go build -a -tags 'netgo' -ldflags '-w -linkmode external -extldflags -static' .

image: build
	@docker build -t $(REPO):$(TAG) .

.PHONY: build image deps
