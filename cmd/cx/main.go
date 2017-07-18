package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	homedir "github.com/mitchellh/go-homedir"
	"gopkg.in/urfave/cli.v1"
)

var (
	Version = "dev"
)

var (
	appFlag = cli.StringFlag{
		Name:  "app, a",
		Usage: "app name inferred from current directory if not specified",
	}
	rackFlag = cli.StringFlag{
		Name:  "rack",
		Usage: "rack name",
	}
)

var globalFlags = []cli.Flag{
	appFlag,
	rackFlag,
}

func init() {
	stdcli.DefaultWriter.Tags["bad"] = stdcli.RenderAttributes(160)
	stdcli.DefaultWriter.Tags["debug"] = stdcli.RenderAttributes(208)
	stdcli.DefaultWriter.Tags["dir"] = stdcli.RenderAttributes(246)
	stdcli.DefaultWriter.Tags["env"] = stdcli.RenderAttributes(95)
	stdcli.DefaultWriter.Tags["good"] = stdcli.RenderAttributes(46)
	stdcli.DefaultWriter.Tags["id"] = stdcli.RenderAttributes(246)
	stdcli.DefaultWriter.Tags["log"] = stdcli.RenderAttributes(45)
	stdcli.DefaultWriter.Tags["name"] = stdcli.RenderAttributes(246)
	stdcli.DefaultWriter.Tags["service"] = stdcli.RenderAttributes(33)
	stdcli.DefaultWriter.Tags["url"] = stdcli.RenderAttributes(246)
	stdcli.DefaultWriter.Tags["version"] = stdcli.RenderAttributes(246)

	cliID()
}

const SysExitCode = 255 // Used to indicate an internal system error

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if ee, ok := err.(*cli.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		os.Exit(SysExitCode)
	}
}

func run() error {
	app := stdcli.New()

	app.Name = "cx"
	app.Version = Version
	app.Usage = "convox management tool"
	app.Flags = globalFlags

	stdcli.VersionPrinter(func(c *cli.Context) {
		runVersion(c)
	})

	ch := make(chan error)

	u, err := user.Current()
	if err != nil {
		return err
	}

	if Version != "dev" && u.Uid != "0" {
		go autoUpdate(ch)
	}

	if err := app.Run(os.Args); err != nil {
		return err
	}

	return nil
}

func appName(c *cli.Context, dir string) (string, error) {
	if app := c.String("app"); app != "" {
		return app, nil
	}

	if app := os.Getenv("CONVOX_APP"); app != "" {
		return app, nil
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	return filepath.Base(abs), nil
}

func autoUpdate(ch chan error) {
	home, err := homedir.Dir()
	if err != nil {
		ch <- err
		return
	}

	lock := filepath.Join(home, ".convox", "autoupdate")

	if stat, err := os.Stat(lock); err == nil {
		if stat.ModTime().After(time.Now().Add(-1 * time.Hour)) {
			ch <- nil
			return
		}
	}

	os.MkdirAll(filepath.Dir(lock), 0755)
	ioutil.WriteFile(lock, []byte{}, 0644)

	ex, err := os.Executable()
	if err != nil {
		ch <- err
		return
	}

	v, err := latestVersion("stable")
	if err != nil {
		ch <- err
		return
	}

	exec.Command(ex, "update", v).Start()

	ch <- nil
}

var errMissingProxyEndpoint = errors.New("Rack endpoint was not found, try cx login")

func Rack(c *cli.Context) rack.Rack {
	var endpoint *url.URL

	exit := func(err error) {
		if err != nil {
			fmt.Fprint(os.Stderr, stdcli.Error(err))
			os.Exit(1)
		}
	}

	local, err := url.Parse("https://localhost:5443")
	if err != nil {
		exit(err)
	}

	if os.Getenv("RACK_URL") == "" {
		proxy, err := consoleProxy()
		if err != nil {
			exit(err)
		}

		if proxy != nil {
			rack, err := currentRack(c)
			if err != nil {
				exit(err)
			}

			switch rack {
			case "":
				fmt.Println("No Rack selected, try cx racks. Using local rack")
				fallthrough
			case "local":
				endpoint = local
			default:
				proxy.Path = fmt.Sprintf("racks/%s", rack)
				endpoint = proxy
			}

		} else {
			endpoint = local
		}

		os.Setenv("RACK_URL", endpoint.String())
	}

	r, err := rack.NewFromEnv()
	if err != nil {
		exit(err)
	}

	return r
}

func cliID() (string, error) {
	fn, err := homedir.Expand("~/.convox/id")
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		id, err := types.Key(32)
		if err != nil {
			return "", err
		}

		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			return "", err
		}

		if err := ioutil.WriteFile(fn, []byte(id), 0644); err != nil {
			return "", err
		}

		return id, nil
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func consoleHost() (string, error) {
	fn, err := homedir.Expand("~/.convox/console/host")
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return "ui.convox.com", nil
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func consoleProxy() (*url.URL, error) {
	fn, err := homedir.Expand("~/.convox/console/proxy")
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	u.Scheme = "https"
	return u, nil
}

func currentRack(c *cli.Context) (string, error) {
	if c.GlobalString("rack") != "" {
		return c.GlobalString("rack"), nil
	}

	if c.String("rack") != "" {
		return c.String("rack"), nil
	}

	return shellRack()
}

func shellRack() (string, error) {
	shpid := os.Getppid()

	fn, err := homedir.Expand(fmt.Sprintf("~/.convox/shell/%d/rack", shpid))
	if err != nil {
		return "", err
	}

	_, err = os.Stat(fn)
	if err != nil {
		// no shell config implies "local" rack
		if os.IsNotExist(err) {
			return "local", nil
		}
		return "", err
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func setConsoleHost(host string) error {
	fn, err := homedir.Expand("~/.convox/console/host")
	if err != nil {
		return err
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(fn, []byte(host), 0644); err != nil {
		return err
	}

	return nil
}

func setConsoleProxy(proxy string) error {
	fn, err := homedir.Expand("~/.convox/console/proxy")
	if err != nil {
		return err
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(fn, []byte(proxy), 0644); err != nil {
		return err
	}

	return nil
}

func setShellRack(rack string) error {
	shpid := os.Getppid()

	fn, err := homedir.Expand(fmt.Sprintf("~/.convox/shell/%d/rack", shpid))
	if err != nil {
		return err
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(fn, []byte(rack), 0644); err != nil {
		return err
	}

	return nil
}

func errorExit(fn cli.ActionFunc, code int) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := fn(c); err != nil {
			if _, ok := err.(cli.ExitCoder); !ok {
				return cli.NewExitError(err.Error(), code)
			}
			return err
		}
		return nil
	}
}

func terminalRaw(f *os.File) (*terminal.State, error) {
	return terminal.MakeRaw(int(f.Fd()))
}

func terminalRestore(f *os.File, state *terminal.State) error {
	return terminal.Restore(int(f.Fd()), state)
}
