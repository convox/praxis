package types

type MetricListOptions struct {
	Start    string
	End      string
	Interval string
}

type MetricGetOptions struct {
	Start    string
	End      string
	Interval string
}

var MetricNames = map[string][]string{
	"balancer": []string{
		"request-count",
		"response-time",
		"server-errror-count",
		"user-error-count",
	},
}
