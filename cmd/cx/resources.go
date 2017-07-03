package main

import (
	"fmt"
	"net"
	"net/url"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Usage: "local port",
			Value: "",
		},
	}
	stdcli.RegisterCommand(cli.Command{
		Name:        "resources",
		Description: "list resources",
		Action:      runResources,
		Before:      beforeCmd,
		Flags:       globalFlags,
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "proxy",
				Description: "proxy connections to a resource",
				Usage:       "<name>",
				Action:      runResourcesProxy,
				Flags:       append(flags, globalFlags...),
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

	stdcli.OK()

	u, err := url.Parse(r.Endpoint)
	if err != nil {
		return err
	}

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

		go handleProxyConnection(cn, app, r.Name)
	}
}

func handleProxyConnection(cn net.Conn, app, resource string) error {
	defer cn.Close()

	r, err := Rack.ResourceProxy(app, resource, cn)
	if err != nil {
		return err
	}

	stdcli.OK()

	return helpers.Stream(cn, r)
}
