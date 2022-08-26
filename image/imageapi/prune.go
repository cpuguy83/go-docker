package imageapi

type Prune struct {
	ImagesDeleted  []DeletedImage
	SpaceReclaimed int64
}

type DeletedImage struct {
	Untagged string
	Deleted  string
}
