package diskutil

// DiskInfo mirrors the output format of the command "diskutil info -plist <disk>" to store information about a disk.
type DiskInfo struct {
	AESHardware                                 bool                `plist:"AESHardware"`
	APFSContainerReference                      string              `plist:"APFSContainerReference"`
	APFSPhysicalStores                          []APFSPhysicalStore `plist:"APFSPhysicalStores"`
	Bootable                                    bool                `plist:"Bootable"`
	BusProtocol                                 string              `plist:"BusProtocol"`
	CanBeMadeBootable                           bool                `plist:"CanBeMadeBootable"`
	CanBeMadeBootableRequiresDestroy            bool                `plist:"CanBeMadeBootableRequiresDestroy"`
	Content                                     string              `plist:"Content"`
	DeviceBlockSize                             int                 `plist:"DeviceBlockSize"`
	DeviceIdentifier                            string              `plist:"DeviceIdentifier"`
	DeviceNode                                  string              `plist:"DeviceNode"`
	DeviceTreePath                              string              `plist:"DeviceTreePath"`
	Ejectable                                   bool                `plist:"Ejectable"`
	EjectableMediaAutomaticUnderSoftwareControl bool                `plist:"EjectableMediaAutomaticUnderSoftwareControl"`
	EjectableOnly                               bool                `plist:"EjectableOnly"`
	FreeSpace                                   uint64              `plist:"FreeSpace"`
	GlobalPermissionsEnabled                    bool                `plist:"GlobalPermissionsEnabled"`
	IOKitSize                                   uint64              `plist:"IOKitSize"`
	IORegistryEntryName                         string              `plist:"IORegistryEntryName"`
	Internal                                    bool                `plist:"Internal"`
	LowLevelFormatSupported                     bool                `plist:"LowLevelFormatSupported"`
	MediaName                                   string              `plist:"MediaName"`
	MediaType                                   string              `plist:"MediaType"`
	MountPoint                                  string              `plist:"MountPoint"`
	OS9DriversInstalled                         bool                `plist:"OS9DriversInstalled"`
	OSInternalMedia                             bool                `plist:"OSInternalMedia"`
	ParentWholeDisk                             string              `plist:"ParentWholeDisk"`
	PartitionMapPartition                       bool                `plist:"PartitionMapPartition"`
	RAIDMaster                                  bool                `plist:"RAIDMaster"`
	RAIDSlice                                   bool                `plist:"RAIDSlice"`
	Removable                                   bool                `plist:"Removable"`
	RemovableMedia                              bool                `plist:"RemovableMedia"`
	RemovableMediaOrExternalDevice              bool                `plist:"RemovableMediaOrExternalDevice"`
	SMARTDeviceSpecificKeysMayVaryNotGuaranteed *SmartDeviceInfo    `plist:"SMARTDeviceSpecificKeysMayVaryNotGuaranteed"`
	SMARTStatus                                 string              `plist:"SMARTStatus"`
	Size                                        uint64              `plist:"Size"`
	SolidState                                  bool                `plist:"SolidState"`
	SupportsGlobalPermissionsDisable            bool                `plist:"SupportsGlobalPermissionsDisable"`
	SystemImage                                 bool                `plist:"SystemImage"`
	TotalSize                                   uint64              `plist:"TotalSize"`
	VirtualOrPhysical                           string              `plist:"VirtualOrPhysical"`
	VolumeName                                  string              `plist:"VolumeName"`
	VolumeSize                                  uint64              `plist:"VolumeSize"`
	WholeDisk                                   bool                `plist:"WholeDisk"`
	Writable                                    bool                `plist:"Writable"`
	WritableMedia                               bool                `plist:"WritableMedia"`
	WritableVolume                              bool                `plist:"WritableVolume"`
}

// ContainerInfo expands on DiskInfo to add extra information for APFS Containers.
type ContainerInfo struct {
	DiskInfo
	APFSContainerFree               uint64 `plist:"APFSContainerFree"`
	APFSContainerSize               uint64 `plist:"APFSContainerSize"`
	APFSSnapshot                    bool   `plist:"APFSSnapshot"`
	APFSSnapshotName                string `plist:"APFSSnapshotName"`
	APFSSnapshotUUID                string `plist:"APFSSnapshotUUID"`
	APFSVolumeGroupID               string `plist:"APFSVolumeGroupID"`
	BooterDeviceIdentifier          string `plist:"BooterDeviceIdentifier"`
	DiskUUID                        string `plist:"DiskUUID"`
	Encryption                      bool   `plist:"Encryption"`
	EncryptionThisVolumeProper      bool   `plist:"EncryptionThisVolumeProper"`
	FileVault                       bool   `plist:"FileVault"`
	FilesystemName                  string `plist:"FilesystemName"`
	FilesystemType                  string `plist:"FilesystemType"`
	FilesystemUserVisibleName       string `plist:"FilesystemUserVisibleName"`
	Fusion                          bool   `plist:"Fusion"`
	Locked                          bool   `plist:"Locked"`
	MacOSSystemAPFSEFIDriverVersion uint64 `plist:"MacOSSystemAPFSEFIDriverVersion"`
	RecoveryDeviceIdentifier        string `plist:"RecoveryDeviceIdentifier"`
	Sealed                          string `plist:"Sealed"`
	VolumeAllocationBlockSize       int    `plist:"VolumeAllocationBlockSize"`
	VolumeUUID                      string `plist:"VolumeUUID"`
}

// APFSPhysicalStore represents the physical device usually relating to synthesized virtual devices.
type APFSPhysicalStore struct {
	DeviceIdentifier string `plist:"APFSPhysicalStore"`
}

// SmartDeviceInfo stores SMART information for devices that are SMART-enabled (e.g. device health or problems).
type SmartDeviceInfo struct {
	AvailableSpare          int `plist:"AVAILABLE_SPARE"`
	AvailableSpareThreshold int `plist:"AVAILABLE_SPARE_THRESHOLD"`
	ControllerBusyTime0     int `plist:"CONTROLLER_BUSY_TIME_0"`
	ControllerBusyTime1     int `plist:"CONTROLLER_BUSY_TIME_1"`
	DataUnitsRead0          int `plist:"DATA_UNITS_READ_0"`
	DataUnitsRead1          int `plist:"DATA_UNITS_READ_1"`
	DataUnitsWritten0       int `plist:"DATA_UNITS_WRITTEN_0"`
	DataUnitsWritten1       int `plist:"DATA_UNITS_WRITTEN_1"`
	HostReadCommands0       int `plist:"HOST_READ_COMMANDS_0"`
	HostReadCommands1       int `plist:"HOST_READ_COMMANDS_1"`
	HostWriteCommands0      int `plist:"HOST_WRITE_COMMANDS_0"`
	HostWriteCommands1      int `plist:"HOST_WRITE_COMMANDS_1"`
	MediaErrors0            int `plist:"MEDIA_ERRORS_0"`
	MediaErrors1            int `plist:"MEDIA_ERRORS_1"`
	NumErrorInfoLogEntries0 int `plist:"NUM_ERROR_INFO_LOG_ENTRIES_0"`
	NumErrorInfoLogEntries1 int `plist:"NUM_ERROR_INFO_LOG_ENTRIES_1"`
	PercentageUsed          int `plist:"PERCENTAGE_USED"`
	PowerCycles0            int `plist:"POWER_CYCLES_0"`
	PowerCycles1            int `plist:"POWER_CYCLES_1"`
	PowerOnHours0           int `plist:"POWER_ON_HOURS_0"`
	PowerOnHours1           int `plist:"POWER_ON_HOURS_1"`
	Temperature             int `plist:"TEMPERATURE"`
	UnsafeShutdowns0        int `plist:"UNSAFE_SHUTDOWNS_0"`
	UnsafeShutdowns1        int `plist:"UNSAFE_SHUTDOWNS_1"`
}
