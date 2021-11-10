VERSION=$(shell git describe --always --tags)
COMMITDATE=$(shell git show -s --format=%ci HEAD)
MODPATH=github.com/aws/ec2-macos-utils
T=./cmd/ec2-macos-utils
V=-v
GO=go

GOIMPORTS = goimports -local "$(T)"

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

clean:
	go clean

GO_TEST_FLAGS=-cover

test: T=./...
test:
	$(GO) test $(V) $(GO_TEST_FLAGS) $(T)

imports::
	$(GOIMPORTS) -w .

imports-check::
	$(GOIMPORTS) -l .
