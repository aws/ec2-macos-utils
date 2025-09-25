## ec2-macos-utils watchdog network-health-monitor

monitor network health

### Synopsis

monitor network health with periodic checks.
A sysdiagnose will be collected on first failure, after which the monitor will exit.

This command requires root privileges. Run with sudo if not running as root.

```
ec2-macos-utils watchdog network-health-monitor [flags]
```

### Options

```
  -h, --help                           help for network-health-monitor
      --interval duration              interval between network checks (default 5m0s)
      --output-base-dir string         base directory for sysdiagnose output (default "/private/var/db/ec2-macos-utils/sysdiagnose")
      --startup-delay duration         delay before starting checks (default 5m0s)
      --sysdiagnose-timeout duration   timeout for sysdiagnose collection (default 15m0s)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging output
```

### SEE ALSO

* [ec2-macos-utils watchdog](ec2-macos-utils_watchdog.md)	 - monitor system health

