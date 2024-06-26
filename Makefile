.DEFAULT_GOAL := test

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

build-test-linux:
	@$(MAKE) -C ./filter/traceable/cmd/libtraceable-downloader build-install-image \
	TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION)
	@docker build -f ./_tests/Dockerfile.test \
	--progress plain \
	--build-arg TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION) \
	-t traceable_goagent_test:$(TRACEABLE_GOAGENT_DISTRO_VERSION) .

build-test-linux-no-lib:
	@$(MAKE) -C ./filter/traceable/cmd/libtraceable-downloader build-install-image \
	TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION)
	@docker build -f ./_tests/Dockerfile.no_lib.test \
	--progress plain \
	--build-arg TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION) \
	-t traceable_goagent_no_lib_test:$(TRACEABLE_GOAGENT_DISTRO_VERSION) .

build-test-linux-missing-lib:
	@$(MAKE) -C ./filter/traceable/cmd/libtraceable-downloader build-install-image \
	TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION)
	@docker build -f ./_tests/Dockerfile.missing_lib.test \
	--progress plain \
	--build-arg TRACEABLE_GOAGENT_DISTRO_VERSION=$(TRACEABLE_GOAGENT_DISTRO_VERSION) \
	-t traceable_goagent_missing_lib_test:$(TRACEABLE_GOAGENT_DISTRO_VERSION) .

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
	$(MAKE) build-test-linux-no-lib TRACEABLE_GOAGENT_DISTRO_VERSION=centos_7
	$(MAKE) build-test-linux-missing-lib TRACEABLE_GOAGENT_DISTRO_VERSION=centos_7

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
	grep _examples/ | \
	xargs -I {} bash -c 'if [ -f "{}/main.go" ] ; then cd {} ; echo "=> {}" ; go build -o ./build_example main.go ; rm build_example ; fi'

.PHONY: fmt
fmt:
	gofmt -w -s ./

porto:
	porto -w .

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
