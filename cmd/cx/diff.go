package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	"github.com/convox/praxis/types"
	"github.com/docker/docker/pkg/archive"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "diff",
		Description: "show changes to be promoted",
		Action:      runDiff,
		Before:      beforeCmd,
		Flags:       globalFlags,
	})
}

func runDiff(c *cli.Context) error {
	app, err := appName(c, ".")
	if err != nil {
		return err
	}

	a, err := Rack(c).AppGet(app)
	if err != nil {
		return err
	}

	if a.Release == "" {
		return fmt.Errorf("no releases for app: %s", app)
	}

	rs, err := Rack(c).ReleaseList(app, types.ReleaseListOptions{Count: 1})
	if err != nil {
		return err
	}

	if len(rs) < 1 {
		return fmt.Errorf("no releases for app: %s", app)
	}

	rt := rs[0]

	rc, err := Rack(c).ReleaseGet(app, a.Release)
	if err != nil {
		return err
	}

	stdcli.Startf("fetching <name>%s</name>", rt.Id)

	bdt, err := fetchBuild(Rack(c), app, rt)
	if err != nil {
		return err
	}

	stdcli.OK()

	stdcli.Startf("fetching <name>%s</name>", rc.Id)

	bdc, err := fetchBuild(Rack(c), app, *rc)
	if err != nil {
		return err
	}

	stdcli.OK()

	cmd := exec.Command("git", "diff", bdc, bdt)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	return nil
}

func fetchBuild(r rack.Rack, app string, release types.Release) (string, error) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}

	c, err := r.ObjectFetch(app, fmt.Sprintf("convox/builds/%s/context.tgz", release.Build))
	if err != nil {
		return "", err
	}

	if err := archive.Untar(c, tmp, &archive.TarOptions{Compression: archive.Gzip}); err != nil {
		return "", err
	}

	ep := []string{}

	for k, v := range release.Env {
		ep = append(ep, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(ep)

	env := strings.Join(ep, "\n")

	if err := ioutil.WriteFile(filepath.Join(tmp, ".env"), []byte(env), 0600); err != nil {
		return "", err
	}

	return tmp, nil
}
