package manifest

import (
	"crypto/sha1"
	"fmt"
)

type Service struct {
	Name string

	Build       ServiceBuild
	Certificate string
	Command     ServiceCommand
	Environment []string
	Health      ServiceHealth
	Image       string
	Port        ServicePort
	Resources   []string
	Scale       ServiceScale
	Volumes     []string
}

type Services []Service

type ServiceBuild struct {
	Args []string
	Path string
}

type ServiceCommand struct {
	Development string
	Test        string
	Production  string
}

type ServiceHealth struct {
	Interval int
	Path     string
	Timeout  int
}

type ServicePort struct {
	Port   int
	Scheme string
}

type ServiceScale struct {
	Count  *ServiceScaleCount
	Memory int
}

type ServiceScaleCount struct {
	Min int
	Max int
}

func (s Service) BuildHash() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("build[path=%q, args=%v] image=%q", s.Build.Path, s.Build.Args, s.Image))))
}

func (s Service) GetName() string {
	return s.Name
}

func (s *Service) SetDefaults() error {
	if s.Scale.Count == nil {
		s.Scale.Count = &ServiceScaleCount{Min: 1, Max: 1}
	}

	return nil
}
