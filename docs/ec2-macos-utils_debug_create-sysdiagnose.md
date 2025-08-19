## ec2-macos-utils debug create-sysdiagnose

create sysdiagnose archive

### Synopsis

creates a sysdiagnose archive including logs, system stats,
and other debug data. The resulting archive will be saved in the specified
output directory.

This command requires root privileges. Run with sudo if not running as root.

```
ec2-macos-utils debug create-sysdiagnose [flags]
```

### Options

```
  -h, --help                help for create-sysdiagnose
      --output-dir string   directory where the sysdiagnose archive will be saved (default "/tmp")
      --timeout duration    set the timeout for creation (e.g. 10m, 30m, 1.5h) (default 15m0s)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging output
```

### SEE ALSO

* [ec2-macos-utils debug](ec2-macos-utils_debug.md)	 - debug utilities for EC2 macOS instances

