package imageapi

// Image represents an image from the docker HTTP API.
type Image struct {
	ID          string `json:"Id,omitempty"`
	ParentID    string `json:"ParentId,omitempty"`
	RepoTags    []string
	RepoDigests []string
	Created     int64
	Size        int64
	SharedSize  int64
	VirtualSize int64
	Labels      map[string]string
	Containers  int64
}
