include ./Makefile.os
include ./Makefile.docker
include ./Makefile.binary

PROJECT_NAME ?=  canary
RELEASE_VERSION ?= latest

.PHONY: release
release: release_prepare release_version release_pkg

release_prepare:
	echo $(shell echo $(RELEASE_VERSION) | tr a-z A-Z) > release.version
	rm -rf ./strimzi-canary-$(RELEASE_VERSION)
	rm -f ./strimzi-canary-$(RELEASE_VERSION).tar.gz
	rm -f ./strimzi-canary-$(RELEASE_VERSION).zip
	mkdir ./strimzi-canary-$(RELEASE_VERSION)

release_version:
	echo "Changing Docker image tags in install to :$(RELEASE_VERSION)"
	$(FIND) ./packaging/install -name '*.yaml' -type f -exec $(SED) -i '/image: "\?quay.io\/strimzi\/[a-zA-Z0-9_.-]\+:[a-zA-Z0-9_.-]\+"\?/s/:[a-zA-Z0-9_.-]\+/:$(RELEASE_VERSION)/g' {} \;

release_pkg:
	$(CP) -r ./packaging/install ./
	$(CP) -r ./packaging/install ././strimzi-canary-$(RELEASE_VERSION)/
	tar -z -cf ./strimzi-canary-$(RELEASE_VERSION).tar.gz strimzi-canary-$(RELEASE_VERSION)/
	zip -r ./strimzi-canary-$(RELEASE_VERSION).zip strimzi-canary-$(RELEASE_VERSION)/
	rm -rf ./strimzi-canary-$(RELEASE_VERSION)

.PHONY: all
all: go_build docker_build docker_push

.PHONY: clean
clean: go_clean