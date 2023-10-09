IMAGECMD=podman
IMAGETAG_REGISTRY=docker.io/peterzandbergen/cesourcewrap
IMAGETAG_TAG=v0.0.1

.PHONY: all
all: build image

.PHONY : test
test:
	@echo Executing test...
	CGO_ENABLED=0 go test --timeout 30s  ./...

.PHONY : build
build: test
	CGO_ENABLED=0 go build -o ./bin/source ./cmd/source

.PHONY: image
image: test
	$(IMAGECMD) build --tag $(IMAGETAG_REGISTRY):$(IMAGETAG_TAG) -f docker/Dockerfile .

.PHONY: push
push: image
	$(IMAGECMD) push $(IMAGETAG_REGISTRY):$(IMAGETAG_TAG)