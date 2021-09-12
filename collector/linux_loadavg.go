package collector

import (
	"context"
	"exporter/parser"
	"exporter/util"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(LoadAvgCollector))
}

type LoadAvgCollector struct{}

var (
	loadAvgLabelNames = []string{}
	loadAvgs          = []*prometheus.Desc{
		prometheus.NewDesc("node_load1", "1m load average.", loadAvgLabelNames, nil),
		prometheus.NewDesc("node_load5", "5m load average.", loadAvgLabelNames, nil),
		prometheus.NewDesc("node_load15", "15m load average.", loadAvgLabelNames, nil),
	}
)

func (collector *LoadAvgCollector) GetName() string {
	return "loadavg"
}

func (collector *LoadAvgCollector) Describe(ch chan<- *prometheus.Desc) error {
	for _, loadAvg := range loadAvgs {
		ch <- loadAvg
	}
	return nil
}

func (collector *LoadAvgCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	path := "cat"
	args := []string{"/proc/loadavg"}

	bytes, err := util.ExecCommand(context.Background(), path, args...)
	if err != nil {
		return fmt.Errorf("get loadavg metric failed, error: %s", err.Error())
	}
	stat, err := p.ParseLoadAvgStat(bytes)
	if err != nil {
		return fmt.Errorf("parse loadavg metric failed, error: %s", err.Error())
	}

	for index, value := range stat {
		ch <- prometheus.MustNewConstMetric(loadAvgs[index], prometheus.GaugeValue, value)
	}

	return nil
}
