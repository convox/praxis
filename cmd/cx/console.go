package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/stdcli"
	cli "gopkg.in/urfave/cli.v1"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "login",
		Description: "login into a rack",
		Action:      runLogin,
	})
	stdcli.RegisterCommand(cli.Command{
		Name:        "racks",
		Description: "list of racks available",
		Action:      runRacks,
	})
}

type Login struct {
	Error string
	Proxy string
}

func runLogin(c *cli.Context) error {
	var host string

	if len(c.Args()) < 1 {
		var err error
		host, err = consoleHost()
		if err != nil {
			return stdcli.Error(err)
		}
	} else {
		host = c.Args()[0]
	}

	fmt.Printf("Email: ")

	reader := bufio.NewReader(os.Stdin)
	email, err := reader.ReadString('\n')
	if err != nil {
		return stdcli.Error(err)
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return stdcli.Errorf("Please provide a valid email")
	}

	fmt.Printf("Password: ")

	pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return stdcli.Error(err)
	}

	fmt.Println()
	stdcli.Startf("Authenticating with <name>%s</name>", host)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}

	u := &url.URL{
		Host:   host,
		Path:   "/cli/login",
		Scheme: "https",
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return stdcli.Error(err)
	}

	req.SetBasicAuth(email, string(pass))

	response, err := client.Do(req)
	if err != nil {
		return stdcli.Error(err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return stdcli.Error(err)
	}

	p := Login{}

	if err := json.Unmarshal(data, &p); err != nil {
		return stdcli.Error(fmt.Errorf("login: %s", err.Error()))
	}

	if p.Error != "" {
		return stdcli.Errorf(p.Error)
	}

	if err := setConsoleHost(host); err != nil {
		return stdcli.Error(err)
	}

	if err := setConsoleProxy(p.Proxy); err != nil {
		return stdcli.Error(err)
	}

	stdcli.OK()
	return nil
}

func runRacks(c *cli.Context) error {
	proxy, err := consoleProxy()
	if err != nil {
		return stdcli.Error(err)
	}

	if proxy == "" {
		return stdcli.Errorf("Console host not found, try cx login")
	}

	endpoint, err := url.Parse(proxy)
	if err != nil {
		return stdcli.Error(err)
	}

	racks := []string{}
	err = consoleClient(endpoint).Get("/racks", rack.RequestOptions{}, &racks)
	if err != nil {
		return stdcli.Error(err)
	}

	t := stdcli.NewTable("RACKS")

	for _, r := range racks {
		t.AddRow(r)
	}

	t.Print()

	return nil
}

func consoleClient(u *url.URL) *rack.Client {
	return &rack.Client{Debug: os.Getenv("CONVOX_DEBUG") == "true", Endpoint: u, Version: "dev"}
}
