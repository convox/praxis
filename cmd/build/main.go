package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/docker/docker/pkg/archive"
)

var (
	Rack rack.Rack

	flagApp      string
	flagId       string
	flagManifest string
	flagPush     string
	flagUrl      string

	output bytes.Buffer
	w      io.Writer
)

func init() {
	r, err := rack.NewFromEnv()
	if err != nil {
		panic(err)
	}

	w = io.MultiWriter(os.Stdout, &output)

	Rack = r
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.StringVar(&flagApp, "app", "", "app name")
	fs.StringVar(&flagId, "id", "", "build id")
	fs.StringVar(&flagManifest, "manifest", "convox.yml", "path to manifest")
	fs.StringVar(&flagPush, "push", "", "push after build")
	fs.StringVar(&flagUrl, "url", "", "source url")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fail(err)
	}

	if v := os.Getenv("BUILD_APP"); v != "" {
		flagApp = v
	}

	if v := os.Getenv("BUILD_ID"); v != "" {
		flagId = v
	}

	if v := os.Getenv("BUILD_MANIFEST"); v != "" {
		flagManifest = v
	}

	if v := os.Getenv("BUILD_PUSH"); v != "" {
		flagPush = v
	}

	if v := os.Getenv("BUILD_URL"); v != "" {
		flagUrl = v
	}

	if err := build(); err != nil {
		fail(err)
	}

	if err := release(); err != nil {
		fail(err)
	}
}

func build() error {
	if _, err := Rack.BuildUpdate(flagApp, flagId, types.BuildUpdateOptions{Started: time.Now(), Status: "running"}); err != nil {
		return err
	}

	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	u, err := url.Parse(flagUrl)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "preparing source\n")

	r, err := Rack.ObjectFetch(flagApp, u.Path)
	if err != nil {
		return err
	}

	if err := archive.Untar(r, tmp, &archive.TarOptions{Compression: archive.Gzip}); err != nil {
		return err
	}

	mf := filepath.Join(tmp, flagManifest)

	m, err := manifest.LoadFile(mf)
	if err != nil {
		return err
	}

	// if err := m.Validate([]string{}); err != nil {
	//   return err
	// }

	data, err := ioutil.ReadFile(mf)
	if err != nil {
		return err
	}

	if _, err := Rack.BuildUpdate(flagApp, flagId, types.BuildUpdateOptions{Manifest: string(data)}); err != nil {
		return err
	}

	opts := manifest.BuildOptions{
		Push:   flagPush,
		Root:   tmp,
		Stdout: w,
		Stderr: w,
	}

	if err := m.Build(flagApp, flagId, opts); err != nil {
		return err
	}

	return nil
}

func release() error {
	release, err := Rack.ReleaseCreate(flagApp, types.ReleaseCreateOptions{Build: flagId})
	if err != nil {
		return err
	}

	if _, err := Rack.BuildUpdate(flagApp, flagId, types.BuildUpdateOptions{Ended: time.Now(), Release: release.Id, Status: "complete"}); err != nil {
		return err
	}

	if _, err := Rack.ObjectStore(flagApp, fmt.Sprintf("convox/builds/%s/log", flagId), &output, types.ObjectStoreOptions{}); err != nil {
		return err
	}

	return nil
}

func fail(err error) {
	fmt.Fprintf(w, "build error: %s\n", err)

	if _, err := Rack.BuildUpdate(flagApp, flagId, types.BuildUpdateOptions{Ended: time.Now(), Status: "failed"}); err != nil {
		fmt.Fprintf(w, "error: could not update build: %s\n", err)
		return
	}

	os.Exit(1)
}
