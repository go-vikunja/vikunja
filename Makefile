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

LDFLAGS := -X "code.vikunja.io/api/pkg/version.Version=$(shell git describe --tags --always --abbrev=10 | sed 's/-/+/' | sed 's/^v//' | sed 's/-g/-/')" -X "main.Tags=$(TAGS)"

PACKAGES ?= $(filter-out code.vikunja.io/api/pkg/integrations,$(shell go list -mod=vendor ./... | grep -v /vendor/))
SOURCES ?= $(shell find . -name "*.go" -type f)

TAGS ?=

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

ifeq ($(DRONE_WORKSPACE),'')
	BINLOCATION := $(EXECUTABLE)
else
    BINLOCATION := $(DIST)/binaries/$(EXECUTABLE)-$(VERSION)-linux-amd64
endif

ifeq ($(VERSION),master)
    PKGVERSION := $(shell git describe --tags --always --abbrev=10 | sed 's/-/+/' | sed 's/^v//' | sed 's/-g/-/')
else
    PKGVERSION := $(VERSION)
endif

.PHONY: all
all: build

.PHONY: clean
clean:
	go clean ./...
	rm -rf $(EXECUTABLE) $(DIST) $(BINDATA)

.PHONY: test
test:
	VIKUNJA_SERVICE_ROOTPATH=$(shell pwd) go test $(GOFLAGS) -cover -coverprofile cover.out $(PACKAGES)

.PHONY: test-coverage
test-coverage: test
	go tool cover -html=cover.out -o cover.html

.PHONY: integration-test
integration-test:
	VIKUNJA_SERVICE_ROOTPATH=$(shell pwd) go test $(GOFLAGS) code.vikunja.io/api/pkg/integrations

.PHONY: lint
lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) golang.org/x/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: build
build: $(EXECUTABLE)

.PHONY: generate
generate:
	go generate code.vikunja.io/api/pkg/static

$(EXECUTABLE): $(SOURCES)
	go build $(GOFLAGS) $(EXTRA_GOFLAGS) -tags '$(TAGS)' -ldflags '-s -w $(LDFLAGS)' -o $@

.PHONY: compress-build
compress-build:
	upx -9 $(EXECUTABLE)

.PHONY: release
release: release-dirs release-windows release-linux release-darwin release-copy release-check release-os-package release-zip

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release $(DIST)/zip

.PHONY: release-windows
release-windows:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) src.techknowlogick.com/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out vikunja-$(VERSION) .
ifneq ($(DRONE_WORKSPACE),'')
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-linux
release-linux:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) src.techknowlogick.com/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out vikunja-$(VERSION) .
ifneq ($(DRONE_WORKSPACE),'')
	mv /build/* $(DIST)/binaries
endif

.PHONY: release-darwin
release-darwin:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install $(GOFLAGS) src.techknowlogick.com/xgo; \
	fi
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out vikunja-$(VERSION) .
ifneq ($(DRONE_WORKSPACE),'')
	mv /build/* $(DIST)/binaries
endif

# Compresses all releases made by make release-* but not mips* releases since upx can't handle these.
.PHONY: release-compress
release-compress:
	$(foreach file,$(filter-out $(wildcard $(wildcard $(DIST)/binaries/$(EXECUTABLE)-*mips*)),$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*)), upx -9 $(file);)

.PHONY: release-copy
release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)

.PHONY: release-check
release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

.PHONY: release-os-package
release-os-package:
	$(foreach file,$(filter-out %.sha256,$(wildcard $(DIST)/release/$(EXECUTABLE)-*)),mkdir $(file)-full;mv $(file) $(file)-full/;	mv $(file).sha256 $(file)-full/; cp config.yml.sample $(file)-full/config.yml; cp LICENSE $(file)-full/; )

.PHONY: release-zip
release-zip:
	$(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),cd $(file); zip -r ../../zip/$(shell basename $(file)).zip *; cd ../../../; )

# Builds a deb package using fpm from a previously created binary (using make build)
.PHONY: build-deb
build-deb:
	fpm -s dir -t deb --url https://vikunja.io -n vikunja -v $(PKGVERSION) --license GPLv3 --directories /opt/vikunja --after-install ./build/after-install.sh --description 'Vikunja is an open-source todo application, written in Go. It lets you create lists,tasks and share them via teams or directly between users.' -m maintainers@vikunja.io ./$(BINLOCATION)=/opt/vikunja/vikunja ./config.yml.sample=/etc/vikunja/config.yml;

.PHONY: reprepro
reprepro:
	reprepro_expect debian includedeb strech ./$(EXECUTABLE)_$(PKGVERSION)_amd64.deb

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
	swag init -g pkg/routes/routes.go -o ./pkg/swagger;
	# Fix the generated swagger file, currently a workaround until swaggo can properly use go mod
	sed -i '/"definitions": {/a "code.vikunja.io.web.HTTPError": {"type": "object","properties": {"code": {"type": "integer"},"message": {"type": "string"}}},' pkg/swagger/docs.go;
	sed -i 's/code.vikunja.io\/web.HTTPError/code.vikunja.io.web.HTTPError/g' pkg/swagger/docs.go;
	sed -i 's/package\ docs/package\ swagger/g' pkg/swagger/docs.go;
	sed -i 's/` + \\"`\\" + `/` + "`" + `/g' pkg/swagger/docs.go;

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
	    go get -u github.com/fzipp/gocyclo; \
		go install $(GOFLAGS) github.com/fzipp/gocyclo; \
	fi
	for S in $(GOFILES); do gocyclo -over 24 $$S || exit 1; done;

.PHONY: static-check
static-check:
	@hash staticcheck > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
	    go get -u honnef.co/go/tools; \
		go install $(GOFLAGS) honnef.co/go/tools/cmd/staticcheck; \
	fi
	staticcheck $(PACKAGES);

.PHONY: gosec-check
gosec-check:
	@hash ./bin/gosec > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s 1.2.0; \
	fi
	for S in $(PACKAGES); do ./bin/gosec $$S || exit 1; done;

.PHONY: goconst-check
goconst-check:
	@hash goconst > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/jgautheron/goconst/cmd/goconst; \
		go install $(GOFLAGS) github.com/jgautheron/goconst/cmd/goconst; \
	fi
	for S in $(PACKAGES); do goconst $$S || exit 1; done;
