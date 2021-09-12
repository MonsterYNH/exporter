package collector

import (
	"exporter/parser"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(BootTimeCollector))
}

type BootTimeCollector struct{}

func (collector *BootTimeCollector) GetName() string {
	return "boot_time"
}

func (collector *BootTimeCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- bootTimeDesc
	return nil
}

func (collector *BootTimeCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {

	bootTime, err := p.ParseBootTime()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(bootTimeDesc, prometheus.CounterValue, bootTime)
	return nil
}

var (
	bootTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "", "boot_time_seconds"),
		"Unix time of last boot, including microseconds.",
		nil, nil,
	)
)
