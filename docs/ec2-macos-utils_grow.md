## ec2-macos-utils grow

resize container to max size

### Synopsis

grow resizes the container to its maximum size using
'diskutil'. The container to operate on can be specified
with its identifier (e.g. disk1 or /dev/disk1). The string
'root' may be provided to resize the OS's root volume.

```
ec2-macos-utils grow [flags]
```

### Options

```
      --dry-run            run command without mutating changes
  -h, --help               help for grow
      --id string          container identifier to be resized or "root"
      --timeout duration   Set the timeout for the command (e.g. 30s, 1m, 1.5h), 0s will disable the timeout (default 5m0s)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging output
```

### SEE ALSO

* [ec2-macos-utils](ec2-macos-utils.md)	 - utilities for EC2 macOS instances

