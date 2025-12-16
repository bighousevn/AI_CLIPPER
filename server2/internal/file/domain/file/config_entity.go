package file

type VideoConfig struct {
	Prompt       string `json:"prompt"`
	ClipCount    int    `json:"clip_count"`
	TargetWidth  int    `json:"target_width"`
	TargetHeight int    `json:"target_height"`
	Subtitle     bool   `json:"subtitle"`
}
