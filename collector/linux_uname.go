package collector

import (
	"exporter/parser"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(UnameCollector))
}

type UnameCollector struct{}

func (collector *UnameCollector) GetName() string {
	return "uname"
}

func (collector *UnameCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *UnameCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	unameStat, err := p.ParseUname()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(unameDesc, prometheus.GaugeValue, 1,
		unameStat.SysName,
		unameStat.Release,
		unameStat.Version,
		unameStat.Machine,
		unameStat.NodeName,
		unameStat.DomainName,
	)

	return nil
}

var unameDesc = prometheus.NewDesc(
	prometheus.BuildFQName("node", "uname", "info"),
	"Labeled system information as provided by the uname system call.",
	[]string{
		"sysname",
		"release",
		"version",
		"machine",
		"nodename",
		"domainname",
	},
	nil,
)
