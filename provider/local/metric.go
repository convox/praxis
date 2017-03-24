package local

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/convox/praxis/types"
)

func (p *Provider) MetricList(app, namespace string) ([]string, error) {
	if metrics, ok := types.MetricNames[namespace]; ok {
		return metrics, nil
	}

	return []string{}, fmt.Errorf("Namespace %s not found", namespace)
}
func (p *Provider) MetricGet(app, namespace, metric string, opts types.MetricGetOptions) ([]string, error) {
	return []string{}, nil
}

func (p *Provider) collectMetrics() {
	for {
		p.collectDockerStats()
		time.Sleep(30 * time.Second)
	}
}

func (p *Provider) collectDockerStats() {

	apps, err := p.AppList()
	if err != nil {
		Logger.Error(fmt.Errorf("stats apps: %s", err))
		return
	}

	dstats := dockerStats()
	if dstats == nil {
		Logger.Logf("empty docker stats")
		return
	}

	for _, app := range apps {
		appMetrics, ok := p.Metrics["service"][app.Name]
		if !ok {
			appMetrics = types.Metric{}
		}

		pss, err := p.ProcessList(app.Name, types.ProcessListOptions{})
		if err != nil {
			Logger.Error(fmt.Errorf("stats ps: %s", err))
			continue
		}

		for _, ps := range pss {
			if metrics, ok := dstats[ps.Id]; ok {
				for _, m := range metrics {
					ms, ok := appMetrics[m.Metric]
					if !ok {
						ms = types.MetricPoints{}
					}
					ms = append(ms, m)
					appMetrics[m.Metric] = ms
				}
			}
		}

		p.Metrics["service"][app.Name] = appMetrics
	}
}

func dockerStats() map[string]types.MetricPoints {
	out, err := exec.Command("docker", "stats", "--no-stream", "--format", "{{.Container}}:{{.CPUPerc}}:{{.MemPerc}}").CombinedOutput()
	if err != nil {
		Logger.Error(fmt.Errorf("stats docker: %s", err))
		return nil
	}

	buffer := make(map[string]types.MetricPoints)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		text := scanner.Text()

		stats := strings.Split(text, ":")

		id := stats[0]

		cpu, err := strconv.ParseFloat(strings.TrimRight(stats[1], "%"), 32)
		if err != nil {
			Logger.Error(fmt.Errorf("stats cpu %s", err))
			return nil
		}

		mem, err := strconv.ParseFloat(strings.TrimRight(stats[2], "%"), 32)
		if err != nil {
			Logger.Error(fmt.Errorf("stats mem %s", err))
			return nil
		}

		ts := time.Now().Format(sortableTime)

		metrics := types.MetricPoints{
			types.MetricPoint{
				Metric:    "cpu",
				Unit:      "percent",
				Value:     float32(cpu),
				Timestamp: ts,
			},
			types.MetricPoint{
				Metric:    "mem",
				Unit:      "percent",
				Value:     float32(mem),
				Timestamp: ts,
			},
		}

		buffer[id] = metrics
	}

	if err := scanner.Err(); err != nil {
		Logger.Error(fmt.Errorf("stats scanner %s", err))
		return nil
	}

	return buffer
}
