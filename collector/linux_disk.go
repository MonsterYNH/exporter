package collector

import (
	"context"
	"exporter/parser"
	"exporter/util"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(DiskCollector))
}

type DiskCollector struct{}

func (collector *DiskCollector) GetName() string {
	return "disk"
}

func (collector *DiskCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- readsCompletedDesc
	ch <- readsMergedDesc
	ch <- readBytesDesc
	ch <- writesCompletedDesc
	ch <- writesMergedDesc
	ch <- writtenBytesDesc
	ch <- ioTimeSecondsDesc
	ch <- readTimeSecondsDesc
	ch <- writeTimeSecondsDesc
	ch <- ioNowDesc
	ch <- discardsCompletedDesc
	ch <- discardsMergedDesc
	ch <- discardedSectorsDesc
	ch <- diskcardTimeSecondsDesc
	ch <- ioTimeWeightedSecondsDesc
	return nil
}

func (collector *DiskCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	path := "cat"
	args := []string{"/proc/diskstats"}

	bytes, err := util.ExecCommand(context.Background(), path, args...)
	if err != nil {
		return fmt.Errorf("get disk metric failed, error: %s", err.Error())
	}

	stat, err := p.ParseDiskStat(bytes)
	if err != nil {
		return fmt.Errorf("host parse disk metric failed, error: %s", err.Error())
	}

	for dev, stats := range stat {
		if ignoredDevicesPattern.MatchString(dev) {
			log.Println(fmt.Sprintf("[INFO] disk ignoring device %s", dev))
			continue
		}

		for index, valueStr := range stats {
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("[INFO] disk parse device %s failed, error: %s", dev, err.Error()))
			}
			switch index {
			case 0:
				ch <- prometheus.MustNewConstMetric(readsCompletedDesc, prometheus.CounterValue, value, dev)
			case 1:
				ch <- prometheus.MustNewConstMetric(readsMergedDesc, prometheus.CounterValue, value, dev)
			case 2:
				ch <- prometheus.MustNewConstMetric(readBytesDesc, prometheus.CounterValue, value, dev)
			case 3:
				ch <- prometheus.MustNewConstMetric(readTimeSecondsDesc, prometheus.CounterValue, value, dev)
			case 4:
				ch <- prometheus.MustNewConstMetric(writesCompletedDesc, prometheus.CounterValue, value, dev)
			case 5:
				ch <- prometheus.MustNewConstMetric(writesMergedDesc, prometheus.CounterValue, value, dev)
			case 6:
				ch <- prometheus.MustNewConstMetric(writtenBytesDesc, prometheus.CounterValue, value, dev)
			case 7:
				ch <- prometheus.MustNewConstMetric(writeTimeSecondsDesc, prometheus.CounterValue, value, dev)
			case 8:
				ch <- prometheus.MustNewConstMetric(ioNowDesc, prometheus.CounterValue, value, dev)
			case 9:
				ch <- prometheus.MustNewConstMetric(ioTimeSecondsDesc, prometheus.CounterValue, value, dev)
			case 10:
				ch <- prometheus.MustNewConstMetric(ioTimeWeightedSecondsDesc, prometheus.CounterValue, value, dev)
			case 11:
				ch <- prometheus.MustNewConstMetric(discardsCompletedDesc, prometheus.CounterValue, value, dev)
			case 12:
				ch <- prometheus.MustNewConstMetric(discardsMergedDesc, prometheus.CounterValue, value, dev)
			case 13:
				ch <- prometheus.MustNewConstMetric(discardedSectorsDesc, prometheus.CounterValue, value, dev)
			case 14:
				ch <- prometheus.MustNewConstMetric(diskcardTimeSecondsDesc, prometheus.CounterValue, value, dev)
			}
		}
	}

	return nil
}

var (
	diskLabelNames = []string{"device"}
	ignore         = "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"

	ignoredDevicesPattern = regexp.MustCompile(ignore)

	readsCompletedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "reads_completed_total"),
		"The total number reads completed successfully",
		diskLabelNames, nil,
	)

	readsMergedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "reads_merged_total"),
		"The total number of reads merged.",
		diskLabelNames,
		nil,
	)

	readBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "read_bytes_total"),
		"The total number of bytes read successfully",
		diskLabelNames, nil,
	)

	writesCompletedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "writes_completed_total"),
		"The total number of writes completed successfully",
		diskLabelNames, nil,
	)

	writesMergedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "writes_merged_total"),
		"The number of writes merged.",
		diskLabelNames,
		nil,
	)

	writtenBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "written_bytes_total"),
		"The total number of bytes written successfully",
		diskLabelNames, nil,
	)

	ioTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "io_time_seconds_total"),
		"The seconds spent doing I/Os",
		diskLabelNames, nil,
	)

	readTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "read_time_seconds_total"),
		"The total number of seconds spent by all reads",
		diskLabelNames, nil,
	)

	writeTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "write_time_seconds_total"),
		"This is the total number of seconds spent by all writes",
		diskLabelNames, nil,
	)

	ioNowDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "io_now"),
		"The number of I/Os currently in progress.",
		diskLabelNames,
		nil,
	)

	discardsCompletedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "discards_completed_total"),
		"The total number of discards completed successfully.",
		diskLabelNames,
		nil,
	)

	discardsMergedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "discards_merged_total"),
		"The total number of discards merged.",
		diskLabelNames,
		nil,
	)

	discardedSectorsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "discarded_sectors_total"),
		"The total number of sectors discarded successfully.",
		diskLabelNames,
		nil,
	)

	diskcardTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "discard_time_seconds_total"),
		"This is the total number of seconds spent by all discards.",
		diskLabelNames,
		nil,
	)

	// flushRequestsDesc = prometheus.NewDesc(
	// 	prometheus.BuildFQName("node", "disk", "flush_requests_total"),
	// 	"The total number of flush requests completed successfully",
	// 	diskLabelNames,
	// 	nil,
	// )

	// flushRequestsTimeSecondsDesc = prometheus.NewDesc(
	// 	prometheus.BuildFQName("node", "disk", "flush_requests_time_seconds_total"),
	// 	"This is the total number of seconds spent by all flush requests.",
	// 	diskLabelNames,
	// 	nil,
	// )

	ioTimeWeightedSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "disk", "io_time_weighted_seconds_total"),
		"The weighted # of seconds spent doing I/Os.",
		diskLabelNames,
		nil,
	)
)
