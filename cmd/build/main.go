package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
	"github.com/docker/docker/pkg/archive"
)

var (
	Rack rack.Rack

	flagApp      string
	flagAuth     string
	flagId       string
	flagManifest string
	flagPrefix   string
	flagPush     string
	flagStage    int
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
	fs.StringVar(&flagPrefix, "prefix", "", "image prefix")
	fs.StringVar(&flagPush, "push", "", "push after build")
	fs.IntVar(&flagStage, "stage", 0, "release stage")
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

	if v := os.Getenv("BUILD_PREFIX"); v != "" {
		flagPrefix = v
	}

	if v := os.Getenv("BUILD_PUSH"); v != "" {
		flagPush = v
	}

	if v := os.Getenv("BUILD_STAGE"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil {
			fail(err)
		}
		flagStage = s
	}

	if v := os.Getenv("BUILD_URL"); v != "" {
		flagUrl = v
	}

	if err := auth(); err != nil {
		fail(err)
	}

	if err := build(); err != nil {
		fail(err)
	}

	if err := release(); err != nil {
		fail(err)
	}
}

func auth() error {
	registries, err := Rack.RegistryList()
	if err != nil {
		return err
	}

	ar, err := Rack.AppRegistry(flagApp)
	if err != nil {
		return err
	}

	registries = append(registries, *ar)

	for _, r := range registries {
		if r.Username == "" && r.Password == "" {
			continue
		}

		if err := exec.Command("docker", "login", "-u", r.Username, "-p", r.Password, r.Hostname).Run(); err != nil {
			return fmt.Errorf("unable to authenticate with registry: %s: %s", r.Hostname, err)
		}
	}

	return nil
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

	data, err := ioutil.ReadFile(mf)
	if err != nil {
		return err
	}

	if _, err := Rack.BuildUpdate(flagApp, flagId, types.BuildUpdateOptions{Manifest: string(data)}); err != nil {
		return err
	}

	cache, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	ce, err := Rack.ObjectExists(flagApp, "convox/cache/build.tgz")
	if err != nil {
		return err
	}

	if ce {
		fmt.Fprintf(w, "restoring cache\n")

		cr, err := Rack.ObjectFetch(flagApp, "convox/cache/build.tgz")
		if err != nil {
			return err
		}

		defer cr.Close()

		if err := helpers.ExtractTarball(cr, cache); err != nil {
			return err
		}
	}

	opts := manifest.BuildOptions{
		Cache:  cache,
		Push:   flagPush,
		Root:   tmp,
		Stdout: w,
		Stderr: w,
	}

	if err := m.Build(flagPrefix, flagId, opts); err != nil {
		return err
	}

	fmt.Fprintf(w, "saving cache\n")

	tgz, err := helpers.CreateTarball(cache, helpers.TarballOptions{})
	if err != nil {
		return err
	}

	defer tgz.Close()

	if _, err := Rack.ObjectStore(flagApp, "convox/cache/build.tgz", tgz, types.ObjectStoreOptions{}); err != nil {
		return err
	}

	return nil
}

func release() error {
	release, err := Rack.ReleaseCreate(flagApp, types.ReleaseCreateOptions{Build: flagId, Stage: flagStage})
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

	if _, err := Rack.ObjectStore(flagApp, fmt.Sprintf("convox/builds/%s/log", flagId), &output, types.ObjectStoreOptions{}); err != nil {
		fmt.Fprintf(w, "error: could not store build logs: %s\n", err)
		return
	}

	os.Exit(1)
}
