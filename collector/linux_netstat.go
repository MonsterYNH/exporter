package collector

import (
	"exporter/parser"
	"fmt"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(NetStatCollector))
}

type NetStatCollector struct{}

func (collector *NetStatCollector) GetName() string {
	return "net_stat"
}

func (collector *NetStatCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *NetStatCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	netStats, err := p.ParseNetStatInfo()
	if err != nil {
		return err
	}

	for protocol, protocolStats := range netStats {
		for name, value := range protocolStats {
			key := protocol + "_" + name
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in netstats: %w", value, err)
			}
			if !netStatFieldsPattern.MatchString(key) {
				continue
			}
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName("node", "netstat", key),
					fmt.Sprintf("Statistic %s.", protocol+name),
					nil, nil,
				),
				prometheus.UntypedValue, v,
			)
		}
	}
	return nil
}

var (
	netStatFields        = "^(.*_(InErrors|InErrs)|Ip_Forwarding|Ip(6|Ext)_(InOctets|OutOctets)|Icmp6?_(InMsgs|OutMsgs)|TcpExt_(Listen.*|Syncookies.*|TCPSynRetrans)|Tcp_(ActiveOpens|InSegs|OutSegs|OutRsts|PassiveOpens|RetransSegs|CurrEstab)|Udp6?_(InDatagrams|OutDatagrams|NoPorts|RcvbufErrors|SndbufErrors))$"
	netStatFieldsPattern = regexp.MustCompile(netStatFields)
)
