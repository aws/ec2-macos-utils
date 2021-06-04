VERSION=$(shell git describe --always --tags)
COMMITDATE=$(shell git show -s --format=%ci HEAD)
MODPATH="github.com/aws/ec2-macos-utils"

build:
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X '${MODPATH}/cmd.CommitDate=${COMMITDATE}' -X '${MODPATH}/cmd.Version=${VERSION}'"

clean:
	go clean

test: build
	@GOOS=darwin GOARCH=amd64 go test ./... -v -cover
