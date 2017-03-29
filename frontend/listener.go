package frontend

func setupListener(name, subnet string) (string, error) {
	destroyListener(name)
	return createListener(name, subnet)
}
