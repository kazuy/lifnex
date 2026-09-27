IMAGE_URI ?= lifnex:local
PLATFORM ?= linux/amd64

.PHONY: build-image

build-image:
	docker build --platform $(PLATFORM) --tag $(IMAGE_URI) app
