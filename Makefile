.DEFAULT_GOAL := test

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

build-test-linux:
	@$(MAKE) -C ./filters/traceable/cmd/libtraceable-downloader build-install-image \
	TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION)

	@docker build -f ./_tests/Dockerfile.test \
	--build-arg TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION) \
	-t traceable_goagent_test:$(TRACEABLE_GOAGENT_DISTRO_VERSION) .

.PHONY: test-linux
test-linux:
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=debian_10
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=debian_11
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=alpine_3.9
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=alpine_3.10
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=alpine_3.11
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=alpine_3.12
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=centos_7
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=centos_8
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=amazonlinux_2
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=ubuntu_18.04
	$(MAKE) build-test-linux TRACEABLE_GOAGENT_DISTRO_VERSION=ubuntu_20.04

.PHONY: bench
bench:
	go test -v -run - -bench . -benchmem ./...

.PHONY: lint
lint:
	@echo "Running linters..."
	@golangci-lint run ./... && echo "Done."

.PHONY: deps
deps:
	@go get -v -t -d ./...

check-examples:
	@find ./ -type d -print | \
	grep examples/ | \
	xargs -I {} bash -c 'if [ -f "{}/main.go" ] ; then cd {} ; echo "=> {}" ; go build -o ./build_example main.go ; rm build_example ; fi'

.PHONY: fmt
fmt:
	gofmt -w -s ./

.PHONY: tidy
tidy:
	@find . -name "go.mod" \
	| grep go.mod \
	| xargs -I {} bash -c 'dirname {}' \
	| xargs -I {} bash -c 'echo "=> {}"; cd {}; go mod tidy -v; '

.PHONY: install-tools
install-tools: ## Install all the dependencies under the tools module
	$(MAKE) -C ./tools install

.PHONY: check-vanity-import
check-vanity-import:
	@porto --skip-files "version.go" -l .

.PHONY: install-libtraceable-downloader
install-libtraceable-downloader:
	cd ./filters/traceable/cmd/libtraceable-downloader && \
	go mod download && \
	go install github.com/Traceableai/goagent/filters/traceable/cmd/libtraceable-downloader

.PHONY: pull-libtraceable-headers
pull-libtraceable-headers:
	$(go env GOPATH)/bin/libtraceable-downloader pull-library-headers "./filters/traceable"
