SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=-race
VERSION = $(shell cat ./VERSION)
BUILDDATE= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILDCOMMIT= $(shell git rev-parse --short HEAD)
VER_IMPORT=github.com/richardcase/casecardgo/pkg/version
FLAGS=-X $(VER_IMPORT).GitHash=$(BUILDCOMMIT) -X $(VER_IMPORT).BuildDate=$(BUILDDATE) -X $(VER_IMPORT).Version=$(VERSION)

setup: ## Install all the build and lint dependencies
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/golang/dep/...
	go get -u github.com/pierrre/gotestcover
	go get -u golang.org/x/tools/cmd/cover
	dep ensure
	gometalinter --install --update

test: ## Run all the tests
	gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=30s

cover: test ## RUn all the tests and opens the coverage report
	go tool cover -html=coverage.txt

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

lint: ## Run all the linters
	gometalinter \
		--exclude=vendor \
		--skip=pkg/client \
		--disable-all \
		--enable=deadcode \
		--enable=ineffassign \
		--enable=staticcheck \
		--enable=gofmt \
		--enable=goimports \
		--enable=misspell \
		--enable=errcheck \
		--enable=vet \
		--enable=vetshadow \
		--deadline=10m \
		./pkg/...
	gometalinter \
		--exclude=vendor \
		--skip=pkg/client \
		--disable-all \
		--enable=deadcode \
		--enable=ineffassign \
		--enable=staticcheck \
		--enable=gofmt \
		--enable=goimports \
		--enable=misspell \
		--enable=errcheck \
		--enable=vet \
		--enable=vetshadow \
		--deadline=10m \
		./cmd/...

ci: lint test ## Run all the tests and code checks

build: ## Build a beta version
	go build -o prepaid-svc ./cmd/prepaid-svc/.
	go build -o prepaid-projector ./cmd/prepaid-projector/.

build-prod: ## Build the production version
	GOOS=linux CGO_ENABLED=0 go build -a \
		--ldflags '$(FLAGS)' \
		-installsuffix cgo \
		-o prepaid-svc \
		./cmd/prepaid-svc/main.go
	GOOS=linux CGO_ENABLED=0 go build -a \
		--ldflags '$(FLAGS)' \
		-installsuffix cgo \
		-o prepaid-projector \
		./cmd/prepaid-projector/main.go


install: ## Install to $GOPATH/src
	go install ./cmd/...

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build