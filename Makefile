# MODPATH is the module's package path.
MODPATH=github.com/aws/ec2-macos-utils
# T is the matching target used to select modules' packages.
T=$(MODPATH)/...
# MAIN is the main executable package to build.
MAIN=./cmd/ec2-macos-utils
# GO is the executable name or path to Go toolchain.
GO=go
# GOIMPORTS is the command to execute goimports with.
GOIMPORTS=$(GO) run golang.org/x/tools/cmd/goimports -local $(T)
# V is a verbosity modifier for some Go toolchain commands.
V=
# GO_TEST_FLAGS specifies additional flags to pass when running 'go test'
GO_TEST_FLAGS=-cover
# GO_BUILD_FLAGS specifies additional flags to pass when running 'go build'
GO_BUILD_FLAGS=
# GOFILES is a list of source files belonging to this module.
GOFILES=$(shell $(GO) list -f '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}' $(T))

# GOOS is left empty to default to host platform. This statement does not
# control the build's artifacts' setting.
export GOOS=
# GOARCH is left empty to default to host platform. This statement does not
# control the build's artifacts' setting.
export GOARCH=
# CGO_ENABLED is set to enabled by default to enable test options.
export CGO_ENABLED=1

# REVISION is the source control revision ID checked out for build.
REVISION=$(shell git describe --always --tags)
# COMMITDATE is the date string associated with the revision.
COMMITDATE=$(shell git show --no-patch --format='%ci' $(REVISION))

# go_ldflags provides build time data to the Go toolchain for the executable.
go_ldflags="-s -w \
            -X '$(MODPATH)/internal/build.CommitDate=$(COMMITDATE)' \
            -X '$(MODPATH)/internal/build.Version=$(REVISION)'"

# BINS lists the set of executables to build. Each is suffixed by their target
# CPU architecture.
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
	$(GO) build -o $@ $(V) -trimpath -ldflags=$(go_ldflags) $(GO_BUILD_FLAGS) $(MAIN)

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

.PHONY: generate
generate::
	go generate $(V) ./...

.PHONY: test
test:
	$(GO) test $(V) $(GO_TEST_FLAGS) $(T)

.PHONY: imports
imports: $(GOFILES)
	$(GOIMPORTS) -w .

.PHONY: imports-check
imports-check:
	$(GOIMPORTS) -e -l -d .
