package manifest

import (
	"fmt"
	"net/url"
	"strings"
)

type Balancer struct {
	Name string

	Endpoints BalancerEndpoints
}

type Balancers []Balancer

type BalancerEndpoint struct {
	Port string

	Protocol string
	Redirect string
	Target   string
}

type BalancerEndpoints []BalancerEndpoint

func (e *BalancerEndpoint) TargetPort() (string, error) {
	if e.Target == "" {
		return "", fmt.Errorf("no target")
	}

	u, err := url.Parse(e.Target)
	if err != nil {
		return "", err
	}

	return u.Port(), nil
}

func (e *BalancerEndpoint) TargetScheme() (string, error) {
	if e.Target == "" {
		return "", fmt.Errorf("no target")
	}

	u, err := url.Parse(e.Target)
	if err != nil {
		return "", err
	}

	return strings.ToUpper(u.Scheme), nil
}
