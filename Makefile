DIST := dist
IMPORT := git.kolaente.de/konrad/list

SED_INPLACE := sed -i

ifeq ($(OS), Windows_NT)
	EXECUTABLE := list.exe
else
	EXECUTABLE := list
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
		SED_INPLACE := sed -i ''
	endif
endif

GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
GOFMT ?= gofmt -s

GOFLAGS := -i -v
EXTRA_GOFLAGS ?=

LDFLAGS := -X "main.Version=$(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')" -X "main.Tags=$(TAGS)"

PACKAGES ?= $(filter-out git.kolaente.de/konrad/list/integrations,$(shell go list ./... | grep -v /vendor/))
SOURCES ?= $(shell find . -name "*.go" -type f)

TAGS ?=

TMPDIR := $(shell mktemp -d 2>/dev/null || mktemp -d -t 'kasino-temp')

ifeq ($(OS), Windows_NT)
	EXECUTABLE := list.exe
else
	EXECUTABLE := list
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

.PHONY: all
all: build

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf $(EXECUTABLE) $(DIST) $(BINDATA)

.PHONY: test
test:
	go test -cover $(PACKAGES)

required-gofmt-version:
	@go version  | grep -q '\(1.7\|1.8\|1.9\|1.10\)' || { echo "We require go version 1.7, 1.8, 1.9 or 1.10 to format code" >&2 && exit 1; }

.PHONY: lint
lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/golang/lint/golint; \
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
release: release-dirs release-windows release-linux release-darwin release-frontend release-copy release-check release-os-package

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release $(DIST)/zip

.PHONY: release-frontend
release-frontend:
	mv frontend/siteconfig.json frontend/siteconfig.json.old; \
    echo '{"API_URL": "/api/v1/"}' > frontend/siteconfig.json; \
	npm --prefix frontend run build; \
	rm frontend/dist/siteconfig.json; \
	mv frontend/siteconfig.json.old frontend/siteconfig.json;

.PHONY: release-windows
release-windows:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out list-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-linux
release-linux:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out list-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-darwin
release-darwin:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/karalabe/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out list-$(VERSION) .
ifeq ($(CI),drone)
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-copy
release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)
	mkdir $(DIST)/release/frontend
	cp frontend/dist $(DIST)/release/frontend/ -R

.PHONY: release-check
release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)


.PHONY: release-os-package
release-os-package:
	$(foreach file,$(filter-out %.sha256,$(wildcard $(DIST)/release/$(EXECUTABLE)-*)),mkdir $(file)-full;mv $(file) $(file)-full/;	mv $(file).sha256 $(file)-full/; cp config.ini.sample $(file)-full/config.ini; cp $(DIST)/release/frontend $(file)-full/ -R; cp LICENSE $(file)-full/; )
	rm $(DIST)/release/frontend -rf

.PHONY: release-zip
release-zip:
	$(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),cd $(file); zip -r ../../zip/$(shell basename $(file)).zip *; cd ../../../; )

.PHONY: generate-swagger
generate-swagger:
	@hash swagger > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/go-swagger/go-swagger/cmd/swagger; \
	fi
	swagger generate spec -o ./public/swagger.v1.json

.PHONY: swagger-check
swagger-check: generate-swagger
	@diff=$$(git diff public/swagger.v1.json); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make generate-swagger' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: swagger-validate
swagger-validate:
	@hash swagger > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/go-swagger/go-swagger/cmd/swagger; \
	fi
	swagger validate ./public/swagger.v1.json
