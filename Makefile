BUILD ?= cmd/example
BUILDDIR ?= build
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

.PHONY: run 
run: vendor
	go run $(BUILD).go

.PHONY: build
build: build-dir vendor
	GOOS=$(OS) GOARCH=$(ARCH) go build -mod vendor -o $(BUILDDIR)/example $(BUILD).go

build-dir:
	@mkdir -p $(BUILDDIR)

.PHONY: clean
clean: clean-output
	@rm -rf $(BUILDDIR)
	@rm -rf vendor


.PHONY: test
test: vendor clean-output
	go test ./... -v

.PHONY: vendor
vendor:
	go mod vendor

.PHONY:
clean-output:
	@rm -f /tmp/job-worker-output-*