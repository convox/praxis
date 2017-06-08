package types

type Image struct {
	Name string

	Args    map[string]string
	Version string
}

type Images []Image

type ImageCreateOptions struct {
	Args    map[string]string
	Version string
}
