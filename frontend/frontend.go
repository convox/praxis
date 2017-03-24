package frontend

import "github.com/convox/logger"

var (
	Log = logger.New("ns=frontend")
)

func Serve(iface, subnet string) error {
	log := Log.At("serve").Namespace("interface=%s subnet=%q", iface, subnet)

	ip, err := setupListener(iface, subnet)
	if err != nil {
		log.Error(err)
		return err
	}

	go startDns(ip)
	go startApi(ip, iface, subnet)

	log.Success()

	select {}

	return nil
}
