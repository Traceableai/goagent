.DEFAULT_GOAL := build-install-images

GO_VERSION ?= 1.16.8

TRACEABLE_GOAGENT_DISTRO = $(firstword $(subst _, ,$(TRACEABLE_GOAGENT_DISTRO_VERSION))) # alpine_1.9 -> alpine
TRACEABLE_GOAGENT_VERSION = $(lastword $(subst _, ,$(TRACEABLE_GOAGENT_DISTRO_VERSION))) # alpine_1.9 -> 1.9

build-install-image:
	docker build -f Dockerfile.$(shell echo $(TRACEABLE_GOAGENT_DISTRO)).test \
	--build-arg $(shell echo $(TRACEABLE_GOAGENT_DISTRO) | tr '[:lower:]' '[:upper:]')_VERSION=$(TRACEABLE_GOAGENT_VERSION) \
	--build-arg GO_VERSION=$(GO_VERSION) \
	-t traceable_goagent_test_base:$(TRACEABLE_GOAGENT_DISTRO_VERSION) .
