package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/convox/praxis/cmd/build/source"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/provider"
	"github.com/convox/praxis/provider/local"
)

var (
	flagApp      string
	flagAuth     string
	flagCache    string
	flagId       string
	flagManifest string
	flagMethod   string
	flagPush     string
	flagRelease  string
	flagUrl      string
)

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.StringVar(&flagAuth, "auth", "", "docker auth data (base64 encoded)")
	fs.StringVar(&flagCache, "cache", "true", "use docker cache")
	fs.StringVar(&flagId, "id", "", "build id")
	fs.StringVar(&flagManifest, "manifest", "convox.yml", "path to manifest")
	fs.StringVar(&flagMethod, "method", "", "source method")
	fs.StringVar(&flagPush, "push", "", "push to registry")
	fs.StringVar(&flagUrl, "url", "", "source url")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fail(err)
	}

	if v := os.Getenv("BUILD_APP"); v != "" {
		flagApp = v
	}

	if v := os.Getenv("BUILD_AUTH"); v != "" {
		flagAuth = v
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

	if flagId == "" {
		fail(fmt.Errorf("no build id"))
	}

	if err := execute(); err != nil {
		fail(err)
	}

	if err := success(); err != nil {
		fail(err)
	}
}

func execute() error {
	if err := login(); err != nil {
		return err
	}

	dir, err := fetch()
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)

	if err := build(dir); err != nil {
		return err
	}

	return nil
}

func fetch() (string, error) {
	var s source.Source

	switch flagMethod {
	case "git":
		s = &source.SourceGit{flagUrl}
	// case "index":
	//   s = &source.SourceIndex{flagUrl}
	case "tgz":
		s = &source.SourceTgz{flagUrl}
	case "zip":
		s = &source.SourceZip{flagUrl}
	default:
		return "", fmt.Errorf("unknown method: %s", flagMethod)
	}

	var buf bytes.Buffer

	dir, err := s.Fetch(&buf)
	log(strings.TrimSpace(buf.String()))
	if err != nil {
		return "", err
	}

	return dir, nil
}

func login() error {
	var auth map[string]struct {
		Username string
		Password string
	}

	if err := json.Unmarshal([]byte(flagAuth), &auth); err != nil {
		return err
	}

	for host, entry := range auth {
		out, err := exec.Command("docker", "login", "-u", entry.Username, "-p", entry.Password, host).CombinedOutput()
		log(fmt.Sprintf("Authenticating %s: %s", host, strings.TrimSpace(string(out))))
		if err != nil {
			return err
		}
	}

	return nil
}

func build(dir string) error {
	dcy := filepath.Join(dir, flagManifest)

	if _, err := os.Stat(dcy); os.IsNotExist(err) {
		return fmt.Errorf("no such file: %s", flagManifest)
	}

	data, err := ioutil.ReadFile(dcy)
	if err != nil {
		return err
	}

	m, err := manifest.Load(data)
	if err != nil {
		return err
	}

	s := make(chan string)

	go func() {
		for l := range s {
			log(l)
		}
	}()

	defer close(s)

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer os.Chdir(wd)

	if err := os.Chdir(dir); err != nil {
		return err
	}

	builds := map[string][]string{}
	pulls := map[string][]string{}

	for _, s := range m.Services {
		switch {
		case s.Build != "":
			builds[s.Build] = append(builds[s.Build], s.Name)
		case s.Image != "":
			pulls[s.Image] = append(pulls[s.Image], s.Name)
		}
	}

	for dir, tags := range builds {
		id := fmt.Sprintf("%x", sha256.Sum256([]byte(dir)))[0:10]

		cmd := exec.Command("docker", "build", "-t", id, dir)

		cmd.Stdout = m.Prefix("build")
		cmd.Stderr = m.Prefix("build")

		cmd.Stdout.Write([]byte(fmt.Sprintf("building %s\n", dir)))

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build failed")
		}

		for _, tag := range tags {
			if err := exec.Command("docker", "tag", id, tag).Run(); err != nil {
				return fmt.Errorf("could not tag: %s", tag)
			}
		}
	}

	for image, tags := range pulls {
		if err := exec.Command("docker", "pull", image).Run(); err != nil {
			return fmt.Errorf("could not pull: %s", image)
		}

		for _, tag := range tags {
			if err := exec.Command("docker", "tag", image, tag).Run(); err != nil {
				return fmt.Errorf("could not tag: %s", tag)
			}
		}
	}

	if flagPush != "" {
		for _, s := range m.Services {
			local := s.Name
			remote := fmt.Sprintf("%s%s:%s", flagPush, s.Name, flagId)

			if err := exec.Command("docker", "tag", local, remote).Run(); err != nil {
				return err
			}

			if err := exec.Command("docker", "push", remote).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func success() error {
	// release := &client.Release{
	//   App: flagApp,
	// }

	// // TODO use provider.ReleaseFork()

	// if flagRelease != "" {
	//   r, err := currentProvider.ReleaseGet(flagApp, flagRelease)
	//   if err != nil {
	//     return err
	//   }
	//   release = r
	// }

	// release.Build = flagId
	// release.Created = time.Now()
	// release.Id = id("R", 10)
	// release.Manifest = currentBuild.Manifest

	// if err := currentProvider.ReleaseSave(release); err != nil {
	//   return err
	// }

	// url, err := Client.BlobStore(flagApp, fmt.Sprintf("convox/builds/%s/logs", currentBuild.Id), bytes.NewReader([]byte(currentLogs)), client.BlobStoreOptions{})
	// if err != nil {
	//   return err
	// }

	// currentBuild.Ended = time.Now()
	// currentBuild.Logs = url
	// // currentBuild.Release = release.Id
	// currentBuild.Status = "complete"

	// if err := Client.BuildSave(currentBuild); err != nil {
	//   return err
	// }

	return nil
}

func fail(err error) {
	log(fmt.Sprintf("ERROR: %s", err))

	// url, _ := Client.BlobStore(flagApp, fmt.Sprintf("convox/builds/%s/logs", currentBuild.Id), bytes.NewReader([]byte(currentLogs)), client.BlobStoreOptions{})

	// currentBuild.Ended = time.Now()
	// currentBuild.Error = err.Error()
	// currentBuild.Logs = url
	// currentBuild.Status = "failed"

	// if err := currentProvider.BuildSave(currentBuild); err != nil {
	//   fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	// }

	os.Exit(1)
}

func log(line string) {
	fmt.Println(line)
}

func providerFromEnv() provider.Provider {
	switch os.Getenv("PROVIDER") {
	default:
		return local.FromEnv()
	}
}
