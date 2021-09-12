package collector

import (
	"exporter/parser"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(IPCollector))
}

type IPCollector struct{}

var (
	ipDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "host_ip", "count"),
		"Each device's ip.",
		[]string{"ip"}, nil,
	)
)

func (collector *IPCollector) GetName() string {
	return "ip"
}

func (collector *IPCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- ipDesc
	return nil
}

func (collector *IPCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	ipStat, err := p.ParseIPStat()
	if err != nil {
		return err
	}

	for _, ip := range ipStat {
		ch <- prometheus.MustNewConstMetric(prometheus.NewDesc(
			prometheus.BuildFQName("node", "host_ip", "count"),
			"Each device's ip.",
			[]string{"ip"}, nil,
		), prometheus.CounterValue, 1, ip)
	}

	return nil
}
