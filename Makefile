IMAGECMD=podman

.PHONY: all
all: build image


.PHONY : build
build:
	go build -o ./source ./cmd/source

.PHONY: image
image:
	$(IMAGECMD) build --tag cesource -f docker/Dockerfile .

.PHONY: push
push: image
	$(IMAGECMD) push --tag 