.DEFAULT_GOAL := test

.PHONY: test
test:
	@go test -count=1 -v -race -cover ./...

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
	@if [[ "$(porto --skip-files ".*\\.pb\\.go$" -l . | wc -c | xargs)" -ne "0" ]]; then echo "Vanity imports are not up to date" ; exit 1 ; fi

.PHONY: install-libtraceable
install-libtraceable:
	@cd ./filters/blocking/cmd/libtraceable-install && \
	TA_BASIC_AUTH_USER="$(TA_BASIC_AUTH_USER)" \
	TA_BASIC_AUTH_TOKEN="$(TA_BASIC_AUTH_TOKEN)" \
	go run main.go $(LIBTRACEABLE_DESTINATION)
