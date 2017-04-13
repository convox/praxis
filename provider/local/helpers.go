package local

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func coalesce(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}

	return ""
}

func coalescei(ints ...int) int {
	for _, i := range ints {
		if i > 0 {
			return i
		}
	}

	return 0
}

func sudoCmd(args ...string) error {
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed: %s - %s", args[0], strings.TrimSpace(string(out)), err)
	}

	return nil
}
