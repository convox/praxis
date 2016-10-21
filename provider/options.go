package provider

type AppCreateOptions struct {
}

type BlobStoreOptions struct {
	Public bool
}

type BuildCreateOptions struct {
	Cache    bool
	Manifest string
}

type ProcessRunOptions struct {
	Command     []string
	Environment map[string]string
}
