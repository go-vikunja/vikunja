DIST := dist
IMPORT := code.vikunja.io/api

SED_INPLACE := sed -i

ifeq ($(OS), Windows_NT)
	EXECUTABLE := vikunja.exe
else
	EXECUTABLE := vikunja
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
		SED_INPLACE := sed -i ''
	endif
endif

GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
GOFMT ?= gofmt -s

GOFLAGS := -v -mod=vendor
EXTRA_GOFLAGS ?=

LDFLAGS := -X "main.Version=$(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')" -X "main.Tags=$(TAGS)"

PACKAGES ?= $(filter-out code.vikunja.io/api/integrations,$(shell go list -mod=vendor ./... | grep -v /vendor/))
SOURCES ?= $(shell find . -name "*.go" -type f)

TAGS ?=

TMPDIR := $(shell mktemp -d 2>/dev/null || mktemp -d -t 'kasino-temp')

ifeq ($(OS), Windows_NT)
	EXECUTABLE := vikunja.exe
else
	EXECUTABLE := vikunja
endif

ifneq ($(DRONE_TAG),)
	VERSION ?= $(subst v,,$(DRONE_TAG))
else
	ifneq ($(DRONE_BRANCH),)
		VERSION ?= $(subst release/v,,$(DRONE_BRANCH))
	else
		VERSION ?= master
	endif
endif

VERSION := $(shell echo $(VERSION) | sed 's/\//\-/g')

.PHONY: all
all: build

.PHONY: clean
clean:
	go clean ./...
	rm -rf $(EXECUTABLE) $(DIST) $(BINDATA)

.PHONY: test
test:
	VIKUNJA_SERVICE_ROOTPATH=$(shell pwd) go test $(GOFLAGS) -cover -coverprofile cover.out $(PACKAGES)
	go tool cover -html=cover.out -o cover.html

required-gofmt-version:
	@go version  | grep -q '\(1.7\|1.8\|1.9\|1.10\|1.11\)' || { echo "We require go version 1.7, 1.8, 1.9, 1.10 or 1.11 to format code" >&2 && exit 1; }

.PHONY: lint
lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) golang.org/x/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: fmt
fmt: required-gofmt-version
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check: required-gofmt-version
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: install
install: $(wildcard *.go)
	go install -v -tags '$(TAGS)' -ldflags '-s -w $(LDFLAGS)'

.PHONY: build
build: $(EXECUTABLE)

$(EXECUTABLE): $(SOURCES)
	go build $(GOFLAGS) $(EXTRA_GOFLAGS) -tags '$(TAGS)' -ldflags '-s -w $(LDFLAGS)' -o $@

.PHONY: release
release: release-dirs release-windows release-linux release-darwin release-copy release-check release-os-package release-zip

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release $(DIST)/zip

.PHONY: release-windows
release-windows:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out vikunja-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-linux
release-linux:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out vikunja-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-darwin
release-darwin:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out vikunja-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-copy
release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)
	mkdir $(DIST)/release/templates
	cp templates/ $(DIST)/templates/ -R

.PHONY: release-check
release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

.PHONY: release-os-package
release-os-package:
	$(foreach file,$(filter-out %.sha256,$(wildcard $(DIST)/release/$(EXECUTABLE)-*)),mkdir $(file)-full;mv $(file) $(file)-full/;	mv $(file).sha256 $(file)-full/; cp config.yml.sample $(file)-full/config.yml; cp $(DIST)/release/templates $(file)-full/ -R; cp LICENSE $(file)-full/; )

.PHONY: release-zip
release-zip:
	$(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),cd $(file); zip -r ../../zip/$(shell basename $(file)).zip *; cd ../../../; )

.PHONY: got-swag
got-swag: do-the-swag
	@diff=$$(git diff docs/swagger/swagger.json); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make do-the-swag' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: do-the-swag
do-the-swag:
	@hash swag > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/swaggo/swag/cmd/swag; \
	fi
	swag init -g pkg/routes/routes.go;
	# Fix the generated swagger file, currently a workaround until swaggo can properly use go mod
	sed -i '/"definitions": {/a "code.vikunja.io.web.HTTPError": {"type": "object","properties": {"code": {"type": "integer"},"message": {"type": "string"}}},' docs/docs.go;
	sed -i 's/code.vikunja.io\/web.HTTPError/code.vikunja.io.web.HTTPError/g' docs/docs.go;

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/client9/misspell/cmd/misspell; \
	fi
	for S in $(GOFILES); do misspell -error $$S || exit 1; done;

.PHONY: ineffassign-check
ineffassign-check:
	@hash ineffassign > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/gordonklaus/ineffassign; \
	fi
	for S in $(GOFILES); do ineffassign $$S || exit 1; done;

.PHONY: gocyclo-check
gocyclo-check:
	@hash gocyclo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) github.com/fzipp/gocyclo; \
	fi
	for S in $(GOFILES); do gocyclo -over 14 $$S || exit 1; done;
