package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/convox/praxis/types"
)

func (p *Provider) RegistryAdd(hostname, username, password string) (*types.Registry, error) {
	registry := &types.Registry{
		Hostname: hostname,
		Username: username,
		Password: password,
	}

	if err := p.registryStore(registry); err != nil {
		return nil, err
	}

	return registry, nil
}

func (p *Provider) RegistryList() (types.Registries, error) {
	return types.Registries{}, nil
}

func (p *Provider) RegistryRemove(hostname string) error {
	return fmt.Errorf("unimplemented")
}

func registryFromAttributes(hostname string, attrs []*simpledb.Attribute) (*types.Registry, error) {
	registry := &types.Registry{Hostname: hostname}

	for _, attr := range attrs {
		switch *attr.Name {
		case "username":
			registry.Username = *attr.Value
		case "password":
			registry.Password = *attr.Value
		}
	}

	return registry, nil
}

func (p *Provider) registryStore(registry *types.Registry) error {
	domain, err := p.rackResource("Registries")
	if err != nil {
		return err
	}

	_, err = p.SimpleDB().PutAttributes(&simpledb.PutAttributesInput{
		Attributes: []*simpledb.ReplaceableAttribute{
			{Name: aws.String("username"), Value: aws.String(registry.Username)},
			{Name: aws.String("password"), Value: aws.String(registry.Password)},
		},
		DomainName: aws.String(domain),
		ItemName:   aws.String(registry.Hostname),
	})

	return err
}
