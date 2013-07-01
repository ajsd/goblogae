package blog

// TODO(arunjit): put this someplace more common
type Range struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit"`
}
