MODPATH=github.com/aws/ec2-macos-utils
T=./cmd/ec2-macos-utils

GO=go
GOIMPORTS=$(GO) run golang.org/x/tools/cmd/goimports -local $(T)
V=
GO_TEST_FLAGS=-cover
GO_BUILD_FLAGS=

GOFILES=$(shell $(GO) list -f '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}' $(MODPATH)/...)

export GOOS
export GOARCH
export CGO_ENABLED=1

REVISION=$(shell git describe --always --tags)
COMMITDATE=$(shell git show --no-patch --format='%ci' $(REVISION))

go_ldflags="-s -w \
            -X '$(MODPATH)/internal/build.CommitDate=$(COMMITDATE)' \
            -X '$(MODPATH)/internal/build.Version=$(REVISION)'"

BINS=bin/ec2-macos-utils_amd64 bin/ec2-macos-utils_arm64

.PHONY: all
all: build test imports docs

.PHONY: build
build: $(BINS)

bin/ec2-macos-utils_%: GOOS=darwin
bin/ec2-macos-utils_%: GOARCH=$*
bin/ec2-macos-utils_%: CGO_ENABLED=0
bin/ec2-macos-utils_%: $(GOFILES)
	@mkdir -p $(@D)
	$(GO) build -o $@ $(V) -trimpath -ldflags=$(go_ldflags) $(GO_BUILD_FLAGS) $(T)

.PHONY: clean
clean:
	$(GO) clean $(if $(V),-x)
	rm -rf bin

.PHONY: docs
docs: docs/ec2-macos-utils.md $(wildcard docs/*.md)

docs/%.md: GOOS=
docs/%.md: GOARCH=
docs/%.md: $(GOFILES)
	$(GO) run $(MODPATH)/internal/cmd/gen-docs $(@D)

.PHONY: test
test: T=$(MODPATH)/...
test:
	$(GO) test $(V) $(GO_TEST_FLAGS) $(T)

.PHONY: imports
imports: $(GOFILES)
	$(GOIMPORTS) -w .

.PHONY: imports-check
imports-check:
	$(GOIMPORTS) -e -l -d .
