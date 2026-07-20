## ec2-macos-utils system cleanup-state

remove OS state for instance imaging

### Synopsis

removes well-known macOS state from the running system that must not be
carried into an image (AMI) built from this instance.

This removes cached OS state, such as the network interface configuration
cache (NetworkInterfaces.plist), required when creating & after provisioning
derived images from an instance.

This command requires root privileges. Run with sudo if not running as root.

```
ec2-macos-utils system cleanup-state [flags]
```

### Options

```
      --dry-run   run command without mutating changes
  -h, --help      help for cleanup-state
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging output
```

### SEE ALSO

* [ec2-macos-utils system](ec2-macos-utils_system.md)	 - system provisioning & helper utilities

