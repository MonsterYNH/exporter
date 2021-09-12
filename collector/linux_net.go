package collector

import (
	"context"
	"exporter/parser"
	"exporter/util"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(NetCollector))
}

type NetCollector struct{}

func (collector *NetCollector) GetName() string {
	return "net"
}

func (collector *NetCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *NetCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	path := "cat"
	args := []string{"/proc/net/dev"}

	bytes, err := util.ExecCommand(context.Background(), path, args...)
	if err != nil {
		return fmt.Errorf("exec netdev command failed, error: %s", err.Error())
	}

	stat, err := p.ParseNetStat(bytes)
	if err != nil {
		return fmt.Errorf("parse netdev metric failed, error: %s", err.Error())
	}

	metricDescs := map[string]*prometheus.Desc{}
	for netdev, devStats := range stat {
		for key, value := range devStats {
			desc, ok := metricDescs[key]
			if !ok {
				desc = prometheus.NewDesc(
					prometheus.BuildFQName("node", "network", key+"_total"),
					fmt.Sprintf("Network device statistic %s.", key),
					[]string{"device"},
					nil,
				)
				metricDescs[key] = desc
			}
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(value), netdev)
		}
	}

	return nil
}
