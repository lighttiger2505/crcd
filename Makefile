NAME := bhb
VERSION := v0.1.0
REVISION := $(shell git rev-parse --short HEAD)
GOVERSION := $(go version)

SRCS := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.version=$(VERSION)\" -X \"main.revision=$(REVISION)\" -X \"main.goversion=$(GOVERSION)\" "
DIST_DIRS := find * -type d -exec

.PHONY: build
build: $(SRCS)
	go build -o $(NAME) $(LDFLAGS) ./...

.PHONY: install
install: $(SRCS)
	go install
	mv $(GOPATH)/bin/browser-hb $(GOPATH)/bin/bhb

.PHONY: dep
dep:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: ensure
ensure: dep
	$(GOPATH)/bin/dep ensure

.PHONY: test
test:
	go test -v

.PHONY: cover
cover:
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: lint
lint:
	@if [ "`gometalinter ./... --config=lint-config.json | tee /dev/stderr`" ]; then \
		echo "^ - lint err" && echo && exit 1; \
	fi

.PHONY: cross-build
cross-build: ensure
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$$os-$$arch/$(NAME); \
		done; \
	done

.PHONY: dist
dist:
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf $(NAME)-$(VERSION)-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r $(NAME)-$(VERSION)-{}.zip {} \; && \
	cd ..
