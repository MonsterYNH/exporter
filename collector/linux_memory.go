package collector

import (
	"context"
	"exporter/parser"
	"exporter/util"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(MemoryCollector))
}

type MemoryCollector struct{}

func (collector *MemoryCollector) GetName() string {
	return "memory"
}

func (collector *MemoryCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *MemoryCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	path := "cat"
	args := []string{"/proc/meminfo"}

	bytes, err := util.ExecCommand(context.Background(), path, args...)
	if err != nil {
		return fmt.Errorf("exec memory command failed, error: %s", err.Error())
	}

	stat, err := p.ParseMemoryStat(bytes)
	if err != nil {
		return fmt.Errorf("parse memory metric failed, error: %s", err.Error())
	}

	var metricType prometheus.ValueType
	for key, value := range stat {
		if strings.HasSuffix(key, "_total") {
			metricType = prometheus.CounterValue
		} else {
			metricType = prometheus.GaugeValue
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName("node", "memory", key),
				fmt.Sprintf("Memory information field %s.", key),
				nil, nil,
			),
			metricType, value,
		)
	}

	return nil
}
