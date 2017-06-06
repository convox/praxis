package manifest

type Resource struct {
	Name string
	Type string
}

type Resources []Resource

func (r Resource) GetName() string {
	return r.Name
}
