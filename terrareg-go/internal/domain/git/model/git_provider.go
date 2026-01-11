package model

type GitProvider struct {
	ID                int
	Name              string
	BaseURLTemplate   string
	CloneURLTemplate  string
	BrowseURLTemplate string
	GitPathTemplate   string
}
