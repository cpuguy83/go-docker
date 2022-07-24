package containerapi

// Resources contains container's resources (cgroups config, ulimits...)
type Resources struct {
	// Applicable to all platforms
	CPUShares int64 `json:"CpuShares"` // CPU shares (relative weight vs. other containers)
	Memory    int64 // Memory limit (in bytes)
	NanoCPUs  int64 `json:"NanoCpus"` // CPU quota in units of 10<sup>-9</sup> CPUs.

	// Applicable to UNIX platforms
	CgroupParent         string // Parent cgroup.
	BlkioWeight          uint16 // Block IO weight (relative weight vs. other containers)
	BlkioWeightDevice    []*WeightDevice
	BlkioDeviceReadBps   []*ThrottleDevice
	BlkioDeviceWriteBps  []*ThrottleDevice
	BlkioDeviceReadIOps  []*ThrottleDevice
	BlkioDeviceWriteIOps []*ThrottleDevice
	CPUPeriod            int64           `json:"CpuPeriod"`          // CPU CFS (Completely Fair Scheduler) period
	CPUQuota             int64           `json:"CpuQuota"`           // CPU CFS (Completely Fair Scheduler) quota
	CPURealtimePeriod    int64           `json:"CpuRealtimePeriod"`  // CPU real-time period
	CPURealtimeRuntime   int64           `json:"CpuRealtimeRuntime"` // CPU real-time runtime
	CpusetCpus           string          // CpusetCpus 0-2, 0,1
	CpusetMems           string          // CpusetMems 0-2, 0,1
	Devices              []DeviceMapping // List of devices to map inside the container
	DeviceCgroupRules    []string        // List of rule to be added to the device cgroup
	DeviceRequests       []DeviceRequest // List of device requests for device drivers
	KernelMemory         int64           // Kernel memory limit (in bytes)
	KernelMemoryTCP      int64           // Hard limit for kernel TCP buffer memory (in bytes)
	MemoryReservation    int64           // Memory soft limit (in bytes)
	MemorySwap           int64           // Total memory usage (memory + swap); set `-1` to enable unlimited swap
	MemorySwappiness     *int64          // Tuning container memory swappiness behaviour
	OomKillDisable       *bool           // Whether to disable OOM Killer or not
	PidsLimit            *int64          // Setting PIDs limit for a container; Set `0` or `-1` for unlimited, or `null` to not change.
	Ulimits              []*Ulimit       // List of ulimits to be set in the container

	// Applicable to Windows
	CPUCount           int64  `json:"CpuCount"`   // CPU count
	CPUPercent         int64  `json:"CpuPercent"` // CPU percent
	IOMaximumIOps      uint64 // Maximum IOps for the container system drive
	IOMaximumBandwidth uint64 // Maximum IO in bytes per second for the container system drive
}

// Ulimit represents the string name of an linux rlimit along with the hard/soft limits.
type Ulimit struct {
	Name string
	Hard int64
	Soft int64
}
