package manifest

type EnvironmentPair struct {
	Key     string
	Default *string
}

type Environment []EnvironmentPair
