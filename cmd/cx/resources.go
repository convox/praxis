package main

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "resources",
		Description: "list resources",
		Action:      runResources,
		Flags:       []cli.Flag{appFlag},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "proxy",
				Description: "proxy connections to a resource",
				Usage:       "<name>",
				Action:      runResourcesProxy,
				Flags: []cli.Flag{
					appFlag,
					cli.StringFlag{
						Name:  "port, p",
						Usage: "local port",
						Value: "",
					},
				},
			},
		},
	})
}

func runResources(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	rs, err := Rack.ResourceList(app)
	if err != nil {
		return err
	}

	t := stdcli.NewTable("NAME", "TYPE", "ENDPOINT")

	for _, r := range rs {
		t.AddRow(r.Name, r.Type, r.Endpoint)
	}

	t.Print()

	return nil
}

func runResourcesProxy(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	if len(c.Args()) < 1 {
		return stdcli.Usage(c)
	}

	name := c.Args()[0]

	stdcli.Startf("starting proxy to <name>%s</name>", name)

	r, err := Rack.ResourceGet(app, name)
	if err != nil {
		return err
	}

	u, err := url.Parse(r.Endpoint)
	if err != nil {
		return err
	}

	stdcli.OK()

	local := u.Port()

	if p := c.String("port"); p != "" {
		local = p
	}

	stdcli.Startf("listening at <url>localhost:%s</url>", local)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", local))
	if err != nil {
		return err
	}

	defer l.Close()

	stdcli.OK()

	uc := *u
	uc.Host = fmt.Sprintf("localhost:%s", local)

	stdcli.Writef("connect to: <url>%s</url>\n\n", &uc)

	for {
		cn, err := l.Accept()
		if err != nil {
			return err
		}

		stdcli.Startf("connection from <url>%s</url>", cn.RemoteAddr())

		go handleProxyConnection(cn, u)
	}
}

func handleProxyConnection(cn net.Conn, target *url.URL) error {
	defer cn.Close()

	pi, err := strconv.Atoi(target.Port())
	if err != nil {
		return err
	}

	r, err := Rack.SystemProxy(target.Hostname(), pi, cn)
	if err != nil {
		return err
	}

	stdcli.OK()

	return helpers.Stream(cn, r)
}
