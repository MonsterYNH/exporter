package collector

import (
	"exporter/parser"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(FileFDStatCollector))
}

type FileFDStatCollector struct{}

func (collector *FileFDStatCollector) GetName() string {
	return "file_fd"
}

func (collector *FileFDStatCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *FileFDStatCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	fileFDStat, err := p.ParseFileFDStat()
	if err != nil {
		return err
	}

	for name, value := range fileFDStat {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid value %s in file-nr: %w", value, err)
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName("node", "filefd", name),
				fmt.Sprintf("File descriptor statistics: %s.", name),
				nil, nil,
			),
			prometheus.GaugeValue, v,
		)
	}
	return nil
}
