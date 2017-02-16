package rack

import (
	"fmt"
	"net/http"
)

type Error struct {
	Error string `json:"error"`
}

func responseError(res *http.Response) error {
	if !res.ProtoAtLeast(2, 0) {
		return fmt.Errorf("server did not respond with http/2")
	}

	if res.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("response status %d", res.StatusCode)
}
