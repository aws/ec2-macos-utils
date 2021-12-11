# EC2 macOS Utils

## Overview

**EC2 macOS Utils** is a CLI-based utility that provides commands for customizing AWS EC2 [Mac instances](https://aws.amazon.com/ec2/instance-types/mac/).
Currently, there exists one command (`grow`) for resizing volumes to their maximum size.
This is done by wrapping `diskutil(8)`, gathering disk information, and resizing the disk.

## Usage

The utility will be installed by default for all AMIs vended by AWS after December 8, 2021.
`ec2-macos-utils` is also available as a cask for install and updates via  [AWS' Homebrew tap](https://github.com/aws/homebrew-aws).

See the [ec2-macos-utils docs](docs/ec2-macos-utils.md) for more information.

### Global Flags

EC2 macOS Utils supports global flags that can be set with any command.
The supported global flags are as follows:
* `--verbose` or `-v` this flag enables more detailed information to be outputted.

### Growing APFS Containers

```
ec2-macos-utils grow [flags]
```

The `grow` command resizes an APFS container to its maximum size.
This is done by fetching all disk and system partition information, repairing the physical device to update partition information, calculating the amount of free space available, and resizing the container to its max size.
Repairing the physical device is necessary in order to properly allocate the amount of available free space.

The `grow` command should be run with `sudo` as it requires root access in order to repair the physical disk.

See the [grow docs](docs/ec2-macos-utils_grow.md) for more information.

## Building

`ec2-macos-utils` can be built using the provided [Makefile](Makefile).
The default target compiles the executable with other targets provided for development & release activities.
Each target has preset values that it uses but these can be overridden and tailored as needed.

### Dependencies

This project requires the following build time dependencies:

- Go
- GNU Make

### Build

```shell
make build
# or, alternatively use the default target to build:
make
```

This builds the `ec2-macos-utils` binary.

### Generate Docs

```shell
make docs
```

This builds the `ec2-macos-utils-docs` binary which can be used to generate the latest [command docs](docs).

### Tests

```shell
make test
```

This runs a cover of all Go tests in the package.

### Imports

```shell
make imports
```

Run [`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to reformat source files according to gofmt.

## Contributing

Please feel free to submit issues, fork the repository and send pull requests!
See [CONTRIBUTING](CONTRIBUTING.md) for more information.

## Security

See the Security section of [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the Apache-2.0 License.