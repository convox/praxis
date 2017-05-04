package types

type Resource struct {
	Name string `json:"name"`

	Endpoint string `json:"endpoint"`
	Type     string `json:"type"`
}

type Resources []Resource
