package manifest

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var (
	Writer io.Writer = os.Stdout
)

type Manifest struct {
	Services Services `yaml:"services,omitempty"`

	Queues Queues `yaml:"queues,omitempty"`
	Tables Tables `yaml:"tables,omitempty"`

	Output io.Writer `yaml:"-"`
}

func Load(data []byte) (*Manifest, error) {
	var m Manifest

	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func LoadFile(filename string) (*Manifest, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Load(data)
}

func Write(data []byte) (int, error) {
	return Writer.Write(data)
}

func Writef(format string, args ...interface{}) (int, error) {
	return Writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (m *Manifest) Raw() ([]byte, error) {
	return yaml.Marshal(m)
}
