package containerapi

// PortMap is a collection of PortBinding indexed by Port
type PortMap map[string][]PortBinding

// PortBinding represents a binding between a Host IP address and a Host Port
type PortBinding struct {
	// HostIP is the host IP Address
	HostIP string `json:"HostIp"`
	// HostPort is the host port number
	HostPort string
}

type LogConfig struct {
	Type   string
	Config map[string]string
}

type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// NetworkingConfig represents the container's networking configuration for each of its interfaces
// Carries the networking configs specified in the `docker run` and `docker network connect` commands
type NetworkingConfig struct {
	EndpointsConfig map[string]*EndpointSettings // Endpoint configs for each connecting network
}

// EndpointSettings stores the network endpoint details
type EndpointSettings struct {
	// Configurations
	IPAMConfig *EndpointIPAMConfig
	Links      []string
	Aliases    []string
	// Operational data
	NetworkID           string
	EndpointID          string
	Gateway             string
	IPAddress           string
	IPPrefixLen         int
	IPv6Gateway         string
	GlobalIPv6Address   string
	GlobalIPv6PrefixLen int
	MacAddress          string
	DriverOpts          map[string]string
}

// EndpointIPAMConfig represents IPAM configurations for the endpoint
type EndpointIPAMConfig struct {
	IPv4Address  string   `json:",omitempty"`
	IPv6Address  string   `json:",omitempty"`
	LinkLocalIPs []string `json:",omitempty"`
}
