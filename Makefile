.DEFAULT_GOAL := test

LIBTRACEABLE_DOWNLOADER ?= libtraceable-downloader

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

.PHONY: test-linux
test-linux:
	@docker build -f Dockerfile.test -t goagent-test .

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
	find ./ -type d -print | \
	grep examples/ | \
	xargs -I {} bash -c 'if [ -f "{}/main.go" ] ; then cd {} ; echo "Building {}" ; go build -o ./build_example main.go ; rm build_example ; fi'

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
	@porto -l .

.PHONY: install-libtraceable-downloader
install-libtraceable-downloader:
	cd ./filters/blocking/cmd/libtraceable-downloader && \
	go mod download && \
	go install github.com/Traceableai/goagent/filters/blocking/cmd/libtraceable-downloader

.PHONY: install-libtraceable
install-libtraceable:
	@$(LIBTRACEABLE_DOWNLOADER) install-library $(LIBTRACEABLE_OS) $(LIBTRACEABLE_DESTINATION)

.PHONY: pull-libtraceable-headers
pull-libtraceable-headers:
	@$(LIBTRACEABLE_DOWNLOADER) pull-library-headers "./filters/blocking/library"
