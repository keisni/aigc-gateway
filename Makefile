VERSION ?= 1.4
image_registry=harbor.zulong.com/common-images
gomod_on = GO111MODULE=on
go_build_cmd = go build
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
build_path := $(dir $(mkfile_path))
project_path := $(realpath $(build_path))

BIN_DIR=$(build_path)/bin
GO_BUILD_LINUX_AMD64=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(gomod_on) $(go_build_cmd)
GO_BUILD_TAGS ?= -gcflags="all=-N -l"

init:
	mkdir -p ${BIN_DIR}

binary: init
	$(GO_BUILD_LINUX_AMD64) \
	-tags $(GO_BUILD_TAGS) -o ${BIN_DIR}/aigc-gateway \
	-installsuffix cgo $(project_path)

base-image:
	docker build $(build_path) --network=host -f Dockerfile.base --tag=$(image_registry)/aigc-base:$(VERSION)

sd-image:
	docker build $(build_path) \
		--build-arg http_proxy=http://10.236.12.73:8118 --build-arg https_proxy=http://10.236.12.73:8118 \
		-f Dockerfile.sd --tag=$(image_registry)/sd-webui:$(VERSION)

image:
	docker build $(build_path) -f Dockerfile --tag=$(image_registry)/aigc-gateway:$(VERSION)

push-base:
	docker push $(image_registry)/aigc-base:$(VERSION)

push-sd:
	docker push $(image_registry)/sd-webui:$(VERSION)

push:
	docker push $(image_registry)/aigc-gateway:$(VERSION)