VERSION=$(shell git describe --always --tags)
COMMITDATE=$(shell git show -s --format=%ci HEAD)
MODPATH=github.com/aws/ec2-macos-utils
T=./cmd/ec2-macos-utils
V=-v
GO=go

GOIMPORTS=golang.org/x/tools/cmd/goimports

export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1

buildpkg=$(MODPATH)/internal/build

go_ldflags="-s -w \
            -X '$(buildpkg).CommitDate=$(COMMITDATE)' \
            -X '$(buildpkg).Version=$(VERSION)'"

build: CGO_ENABLED=0
build:
	$(GO) build $(V) -trimpath -ldflags=$(go_ldflags) $(GO_BUILD_FLAGS) $(T)

clean: GO_CLEAN_FLAGS=-x
clean:
	$(GO) clean $(GO_CLEAN_FLAGS)
	rm -f ec2-macos-utils
	rm -f ec2-macos-utils-docs

.PHONY: docs
docs: GO_BUILD_FLAGS=-tags docs
docs: T=./cmd/ec2-macos-utils-docs
docs:
	$(GO) build $(V) $(GO_BUILD_FLAGS) $(T)

GO_TEST_FLAGS=-cover

test: T=./...
test:
	$(GO) test $(V) $(GO_TEST_FLAGS) $(T)

imports::
	$(GO) run $(GOIMPORTS) -local "$(T)" -w .

imports-check::
	$(GO) run $(GOIMPORTS) -local "$(T)" -l .
