.DEFAULT_GOAL := build-install-images

GO_VERSION ?= 1.16.8

build-install-images:
	docker build -f Dockerfile.alpine.test \
	--build-arg ALPINE_VERSION=3.9 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_alpine_3.9:dev .

	docker build -f Dockerfile.alpine.test \
	--build-arg ALPINE_VERSION=3.12 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_alpine_3.12:dev .

	docker build -f Dockerfile.centos.test \
	--build-arg CENTOS_VERSION=7 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_centos_7:dev .

	docker build -f Dockerfile.centos.test \
	--build-arg CENTOS_VERSION=8 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_centos_8:dev .

	docker build -f Dockerfile.amazonlinux_2.test \
	--build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_amazonlinux_2:dev .

	docker build -f Dockerfile.ubuntu.test \
	--build-arg UBUNTU_VERSION=18.04 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_ubuntu_18.04:dev .

	docker build -f Dockerfile.ubuntu.test \
	--build-arg UBUNTU_VERSION=20.04 --build-arg GO_VERSION=1.16.8 \
	-t traceable_goagent_ubuntu_20.04:dev .