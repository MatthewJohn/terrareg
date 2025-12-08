package terrareg

// TerraregCustomLink represents a custom link configuration.
type TerraregCustomLink struct {
	Text string `json:"text"`              // Python uses "text" not "title"
	URL  string `json:"url"`
}