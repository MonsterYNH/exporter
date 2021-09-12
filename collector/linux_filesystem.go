package collector

import (
	"context"
	"errors"
	"exporter/parser"
	"exporter/util"
	"fmt"
	"log"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(FileSystemCollector))
}

type FileSystemCollector struct{}

func (collector *FileSystemCollector) GetName() string {
	return "filesystem"
}

func (collector *FileSystemCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- sizeDesc
	ch <- freeDesc
	ch <- availDesc
	ch <- filesDesc
	ch <- filesFreeDesc
	ch <- roDesc
	ch <- deviceErrorDesc

	return nil
}

func (collector *FileSystemCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	bytes, err := mountpointDetails()
	if err != nil {
		return fmt.Errorf("[ERROR] get mountpoint details failed, error: %s", err.Error())
	}

	stats, err := p.ParseFileSystemStat(bytes)
	if err != nil {
		return fmt.Errorf("[ERROR] parse filesystem metric failed, error: %s", err.Error())
	}

	// Make sure we expose a metric once, even if there are multiple mounts
	seen := map[parser.FileSystemLabels]bool{}
	for _, s := range stats {
		if seen[s.Labels] {
			continue
		}
		seen[s.Labels] = true

		ch <- prometheus.MustNewConstMetric(
			deviceErrorDesc, prometheus.GaugeValue,
			s.DeviceError, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		if s.DeviceError > 0 {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			sizeDesc, prometheus.GaugeValue,
			s.Size, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		ch <- prometheus.MustNewConstMetric(
			freeDesc, prometheus.GaugeValue,
			s.Free, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		ch <- prometheus.MustNewConstMetric(
			availDesc, prometheus.GaugeValue,
			s.Avail, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		ch <- prometheus.MustNewConstMetric(
			filesDesc, prometheus.GaugeValue,
			s.Files, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		ch <- prometheus.MustNewConstMetric(
			filesFreeDesc, prometheus.GaugeValue,
			s.FilesFree, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
		ch <- prometheus.MustNewConstMetric(
			roDesc, prometheus.GaugeValue,
			s.Ro, s.Labels.Device, s.Labels.MountPoint, s.Labels.FsType,
		)
	}
	return nil
}

func mountpointDetails() ([]byte, error) {
	var (
		path         = "cat"
		args         = "/proc/1/mounts"
		fallbackArgs = "/proc/mounts"
	)

	bytes, err := util.ExecCommand(context.Background(), path, args)
	if errors.Is(err, os.ErrNotExist) {
		log.Println("Reading root mounts failed, falling back to system mounts" + err.Error())
		bytes, err = util.ExecCommand(context.Background(), path, fallbackArgs)
	}

	return bytes, err
}

var (
	filesystemLabelNames = []string{"device", "mountpoint", "fstype"}
	sizeDesc             = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "size_bytes"),
		"Filesystem size in bytes.",
		filesystemLabelNames, nil,
	)

	freeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "free_bytes"),
		"Filesystem free space in bytes.",
		filesystemLabelNames, nil,
	)

	availDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "avail_bytes"),
		"Filesystem space available to non-root users in bytes.",
		filesystemLabelNames, nil,
	)

	filesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "files"),
		"Filesystem total file nodes.",
		filesystemLabelNames, nil,
	)

	filesFreeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "files_free"),
		"Filesystem total free file nodes.",
		filesystemLabelNames, nil,
	)

	roDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "readonly"),
		"Filesystem read-only status.",
		filesystemLabelNames, nil,
	)

	deviceErrorDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "filesystem", "device_error"),
		"Whether an error occurred while getting statistics for the given device.",
		filesystemLabelNames, nil,
	)
)
