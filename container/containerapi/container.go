package containerapi

import (
	"time"

	"github.com/cpuguy83/go-docker/container/containerapi/mount"
)

// Container represents a container
type Container struct {
	ID              string `json:"Id"`
	Names           []string
	Created         int
	Path            string
	Args            []string
	State           string
	Image           string
	ImageID         string
	Command         string
	Ports           []NetworkPort
	Labels          map[string]string
	HostConfig      *HostConfig
	NetworkSettings *NetworkSettings
	Mounts          []MountPoint
	SizeRootFs      int
	SizeRw          int
}

// ContainerInspect is newly used struct along with MountPoint
type ContainerInspect struct {
	ID              string `json:"Id"`
	Created         string
	Path            string
	Args            []string
	State           *ContainerState
	Image           string
	ResolvConfPath  string
	HostnamePath    string
	HostsPath       string
	LogPath         string
	Name            string
	RestartCount    int
	Driver          string
	Platform        string
	MountLabel      string
	ProcessLabel    string
	AppArmorProfile string
	ExecIDs         []string
	HostConfig      *HostConfig
	GraphDriver     GraphDriverData
	SizeRw          *int64 `json:",omitempty"`
	SizeRootFs      *int64 `json:",omitempty"`
	Mounts          []MountPoint
	Config          *Config
	NetworkSettings *NetworkSettings
}

// NetworkAddress represents an IP address
type NetworkAddress struct {
	Addr      string
	PrefixLen int
}

// NetworkPort represents a network port
type NetworkPort struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

// NetworkSettings exposes the network settings in the api
type NetworkSettings struct {
	Bridge                 string  // Bridge is the Bridge name the network uses(e.g. `docker0`)
	SandboxID              string  // SandboxID uniquely represents a container's network stack
	HairpinMode            bool    // HairpinMode specifies if hairpin NAT should be enabled on the virtual interface
	LinkLocalIPv6Address   string  // LinkLocalIPv6Address is an IPv6 unicast address using the link-local prefix
	LinkLocalIPv6PrefixLen int     // LinkLocalIPv6PrefixLen is the prefix length of an IPv6 unicast address
	Ports                  PortMap // Ports is a collection of PortBinding indexed by Port
	SandboxKey             string  // SandboxKey identifies the sandbox
	SecondaryIPAddresses   []NetworkAddress
	SecondaryIPv6Addresses []NetworkAddress
	Networks               map[string]*EndpointSettings
}

// ContainerState stores container's running state
// it's part of ContainerJSONBase and will return by "inspect" command
type ContainerState struct {
	Status     string // String representation of the container state. Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	ExitCode   int
	Error      string
	StartedAt  string
	FinishedAt string
	Health     *Health `json:",omitempty"`
}

// MountPoint represents a mount point configuration inside the container.
// This is used for reporting the mountpoints in use by a container.
type MountPoint struct {
	Type        mount.Type `json:",omitempty"`
	Name        string     `json:",omitempty"`
	Source      string
	Destination string
	Driver      string `json:",omitempty"`
	Mode        string
	RW          bool
	Propagation mount.Propagation
}

// GraphDriverData Information about a container's graph driver.
// swagger:model GraphDriverData
type GraphDriverData struct {

	// data
	// Required: true
	Data map[string]string `json:"Data"`

	// name
	// Required: true
	Name string `json:"Name"`
}

// Health stores information about the container's healthcheck results
type Health struct {
	Status        string               // Status is one of Starting, Healthy or Unhealthy
	FailingStreak int                  // FailingStreak is the number of consecutive failures
	Log           []*HealthcheckResult // Log contains the last few results (oldest first)
}

// HealthcheckResult stores information about a single run of a healthcheck probe
type HealthcheckResult struct {
	Start    time.Time // Start is the time this check started
	End      time.Time // End is the time this check ended
	ExitCode int       // ExitCode meanings: 0=healthy, 1=unhealthy, 2=reserved (considered unhealthy), else=error running probe
	Output   string    // Output from last check
}

// DeviceRequest represents a request for devices from a device driver.
// Used by GPU device drivers.
type DeviceRequest struct {
	Driver       string            // Name of device driver
	Count        int               // Number of devices to request (-1 = All)
	DeviceIDs    []string          // List of device IDs as recognizable by the device driver
	Capabilities [][]string        // An OR list of AND lists of device capabilities (e.g. "gpu")
	Options      map[string]string // Options to pass onto the device driver
}

// DeviceMapping represents the device mapping between the host and the container.
type DeviceMapping struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}
