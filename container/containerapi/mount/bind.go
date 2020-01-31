package mount

// Consistency represents the consistency requirements of a mount.
type Consistency string

const (
	// ConsistencyFull guarantees bind mount-like consistency
	ConsistencyFull Consistency = "consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	ConsistencyCached Consistency = "cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	ConsistencyDelegated Consistency = "delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	ConsistencyDefault Consistency = "default"
)

// BindOptions defines options specific to mounts of type "bind".
type BindOptions struct {
	Propagation  Propagation `json:",omitempty"`
	NonRecursive bool        `json:",omitempty"`
}
