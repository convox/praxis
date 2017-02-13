package manifest

import (
	"crypto/sha1"
	"fmt"
)

type Service struct {
	Name string

	Build       ServiceBuild
	Environment []string
	Image       string
}

type Services []Service

type ServiceBuild struct {
	Args []string
	Path string
}

func (sb ServiceBuild) Hash() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%v||||%v", sb.Path, sb.Args))))
}
