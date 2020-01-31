package mount

// Type represents the type of a mount.
type Type string

// Type constants
const (
	// TypeBind is the type for mounting host dir
	TypeBind Type = "bind"
	// TypeVolume is the type for remote storage volumes
	TypeVolume Type = "volume"
	// TypeTmpfs is the type for mounting tmpfs
	TypeTmpfs Type = "tmpfs"
	// TypeNamedPipe is the type for mounting Windows named pipes
	TypeNamedPipe Type = "npipe"
)

// Mount represents a mount (volume).
type Mount struct {
	Type Type `json:",omitempty"`
	// Source specifies the name of the mount. Depending on mount type, this
	// may be a volume name or a host path, or even ignored.
	// Source is not supported for tmpfs (must be an empty value)
	Source      string      `json:",omitempty"`
	Target      string      `json:",omitempty"`
	ReadOnly    bool        `json:",omitempty"`
	Consistency Consistency `json:",omitempty"`

	BindOptions   *BindOptions   `json:",omitempty"`
	VolumeOptions *VolumeOptions `json:",omitempty"`
	TmpfsOptions  *TmpfsOptions  `json:",omitempty"`
}
