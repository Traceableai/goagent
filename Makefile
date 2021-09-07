.DEFAULT_GOAL := test

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

.PHONY: test-linux
test-linux:
	@docker build -f Dockerfile.test -t goagent-test \
	--build-arg TA_BASIC_AUTH_USER=$(TA_BASIC_AUTH_USER) \
	--build-arg TA_BASIC_AUTH_TOKEN=$(TA_BASIC_AUTH_TOKEN) .

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

.PHONY: ci-deps
deps-ci:
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint

check-examples:
	find ./ -type d -print | \
	grep examples/ | \
	xargs -I {} bash -c 'if [ -f "{}/main.go" ] ; then cd {} ; echo "Building {}" ; go build -o ./build_example main.go ; rm build_example ; fi'

.PHONY: fmt
fmt:
	gofmt -w -s ./

.PHONY: tidy
tidy:
	find . -path ./config -prune -o -name "go.mod" \
	| grep go.mod \
	| xargs -I {} bash -c 'dirname {}' \
	| xargs -I {} bash -c 'cd {}; go mod tidy'

.PHONY: install-tools
install-tools: ## Install all the dependencies under the tools module
	$(MAKE) -C ./tools install

.PHONY: check-vanity-import
check-vanity-import:
	@porto -l .

.PHONY: install-libtraceable-downloader
install-libtraceable-downloader:
	cd ./filters/blocking/cmd/libtraceable-downloader && \
	go mod download && \
	go install github.com/Traceableai/goagent/filters/blocking/cmd/libtraceable-downloader

.PHONY: install-libtraceable
install-libtraceable:
	@libtraceable-downloader install-library $(LIBTRACEABLE_OS) $(LIBTRACEABLE_DESTINATION)

.PHONY: pull-libtraceable-headers
pull-libtraceable-headers:
	@libtraceable-downloader pull-library-headers "./filters/blocking/library"
