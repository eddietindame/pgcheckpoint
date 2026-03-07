.PHONY: build run test tag-patch tag-minor tag-major

build:
	go build -o bin/pgcheckpoint .

run:
	go run . $(ARGS)

test:
	gotestsum

LATEST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION := $(subst v,,$(LATEST_TAG))
MAJOR := $(word 1,$(subst ., ,$(VERSION)))
MINOR := $(word 2,$(subst ., ,$(VERSION)))
PATCH := $(word 3,$(subst ., ,$(VERSION)))

tag-patch:
	@echo "$(LATEST_TAG) -> v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))"
	git tag "v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))"

tag-minor:
	@echo "$(LATEST_TAG) -> v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0"
	git tag "v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0"

tag-major:
	@echo "$(LATEST_TAG) -> v$(shell echo $$(($(MAJOR)+1))).0.0"
	git tag "v$(shell echo $$(($(MAJOR)+1))).0.0"

