package collector

import (
	"context"
	"exporter/parser"
	"exporter/util"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(CPUCollector))
}

type CPUCollector struct{}

var (
	nodeCPUSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "cpu", "seconds_total"),
		"Seconds the CPUs spent in each mode.",
		[]string{"cpu", "mode"}, nil,
	)
)

func (collector *CPUCollector) GetName() string {
	return "cpu"
}

func (collector *CPUCollector) Describe(ch chan<- *prometheus.Desc) error {
	ch <- nodeCPUSecondsDesc
	return nil
}

func (collector *CPUCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	path := "cat"
	args := []string{"/proc/stat"}

	bytes, err := util.ExecCommand(context.Background(), path, args...)
	if err != nil {
		return fmt.Errorf("get cpu metric failed, error: %s", err.Error())
	}
	stat, err := p.ParseCPUStat(bytes)
	if err != nil {
		return fmt.Errorf("parse cpu metric failed, error: %s", err.Error())
	}

	for cpuID, cpuStat := range stat.CPU {
		cpuNum := strconv.Itoa(cpuID)
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.User, cpuNum, "user")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.Nice, cpuNum, "nice")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.System, cpuNum, "system")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.Idle, cpuNum, "idle")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.Iowait, cpuNum, "iowait")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.IRQ, cpuNum, "irq")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.SoftIRQ, cpuNum, "softirq")
		ch <- prometheus.MustNewConstMetric(nodeCPUSecondsDesc, prometheus.CounterValue, cpuStat.Steal, cpuNum, "steal")
	}

	return nil
}
