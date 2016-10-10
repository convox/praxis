package exec

import "os/exec"

type Cmd struct {
	*exec.Cmd
}

func Command(path string, args ...string) Cmd {
	return Cmd{
		Cmd: exec.Command(path, args...),
	}
}
