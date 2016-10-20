package manifest

import "path/filepath"

type Volume struct {
	Local  string
	Remote string
}

type Volumes []Volume

func (vv Volumes) Prepend(path string) {
	for i := range vv {
		vv[i].Local = filepath.Join(path, vv[i].Local)
	}
}
