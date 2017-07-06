package router

func createAlias(iface, ip string) error {
	err := execute("ip", "addr", "add", ip, "dev", iface)
	return err
}
