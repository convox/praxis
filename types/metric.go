package types

type MetricGetOptions struct {
	Start    string
	End      string
	Interval string
}

type MetricPoints []MetricPoint

type MetricPoint struct {
	Metric    string
	Timestamp string
	Unit      string
	Value     float32
}

type MetricNamespace map[string]MetricEntity
type MetricEntity map[string]Metric
type Metric map[string]MetricPoints

var MetricNames = map[string][]string{
	"balancer": []string{
		"request-count",
		"response-time",
		"server-errror-count",
		"user-error-count",
	},
}