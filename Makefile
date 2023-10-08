IMAGECMD=podman
IMAGETAG_REGISTRY=docker.io/peterzandbergen/cesourcewrap
IMAGETAG_TAG=v0.0.1

.PHONY: all
all: build image


.PHONY : build
build:
	go build -o ./source ./cmd/source

.PHONY: image
image:
	$(IMAGECMD) build --tag $(IMAGETAG_REGISTRY):$(IMAGETAG_TAG) -f docker/Dockerfile .

.PHONY: push
push: image
	$(IMAGECMD) push $(IMAGETAG_REGISTRY):$(IMAGETAG_TAG)