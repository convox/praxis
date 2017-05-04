package types

type Service struct {
	Name string `json:"name"`

	Endpoint string `json:"endpoint"`
}

type Services []Service
