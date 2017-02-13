package manifest

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
