package types

import "time"

type LogsOptions struct {
	Filter string
	Follow bool
	Since  time.Time
}
