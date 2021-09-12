package collector

import (
	"exporter/parser"
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector(new(NetClassCollector))
}

type NetClassCollector struct{}

func (collector *NetClassCollector) GetName() string {
	return "net_class"
}

func (collector *NetClassCollector) Describe(ch chan<- *prometheus.Desc) error {
	return nil
}

func (collector *NetClassCollector) Collect(p parser.Parser, ch chan<- prometheus.Metric) error {
	netClass, err := p.ParseNetClass()
	if err != nil {
		return err
	}

	for device, ifaceInfo := range netClass {
		if ignoredNetClassDevicesPattern.MatchString(device) {
			continue
		}

		upDesc := prometheus.NewDesc(
			prometheus.BuildFQName("node", "network", "up"),
			"Value is 1 if operstate is 'up', 0 otherwise.",
			[]string{"device"},
			nil,
		)
		upValue := 0.0
		if ifaceInfo.OperState == "up" {
			upValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, upValue, ifaceInfo.Name)

		infoDesc := prometheus.NewDesc(
			prometheus.BuildFQName("node", "network", "info"),
			"Non-numeric data from /sys/class/net/<iface>, value is always 1.",
			[]string{"device", "address", "broadcast", "duplex", "operstate", "ifalias"},
			nil,
		)
		infoValue := 1.0

		ch <- prometheus.MustNewConstMetric(infoDesc, prometheus.GaugeValue, infoValue, ifaceInfo.Name, ifaceInfo.Address, ifaceInfo.Broadcast, ifaceInfo.Duplex, ifaceInfo.OperState, ifaceInfo.IfAlias)

		if ifaceInfo.AddrAssignType != nil {
			pushMetric(ch, "address_assign_type", *ifaceInfo.AddrAssignType, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.Carrier != nil {
			pushMetric(ch, "carrier", *ifaceInfo.Carrier, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.CarrierChanges != nil {
			pushMetric(ch, "carrier_changes_total", *ifaceInfo.CarrierChanges, ifaceInfo.Name, prometheus.CounterValue)
		}

		if ifaceInfo.CarrierUpCount != nil {
			pushMetric(ch, "carrier_up_changes_total", *ifaceInfo.CarrierUpCount, ifaceInfo.Name, prometheus.CounterValue)
		}

		if ifaceInfo.CarrierDownCount != nil {
			pushMetric(ch, "carrier_down_changes_total", *ifaceInfo.CarrierDownCount, ifaceInfo.Name, prometheus.CounterValue)
		}

		if ifaceInfo.DevID != nil {
			pushMetric(ch, "device_id", *ifaceInfo.DevID, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.Dormant != nil {
			pushMetric(ch, "dormant", *ifaceInfo.Dormant, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.Flags != nil {
			pushMetric(ch, "flags", *ifaceInfo.Flags, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.IfIndex != nil {
			pushMetric(ch, "iface_id", *ifaceInfo.IfIndex, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.IfLink != nil {
			pushMetric(ch, "iface_link", *ifaceInfo.IfLink, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.LinkMode != nil {
			pushMetric(ch, "iface_link_mode", *ifaceInfo.LinkMode, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.MTU != nil {
			pushMetric(ch, "mtu_bytes", *ifaceInfo.MTU, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.NameAssignType != nil {
			pushMetric(ch, "name_assign_type", *ifaceInfo.NameAssignType, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.NetDevGroup != nil {
			pushMetric(ch, "net_dev_group", *ifaceInfo.NetDevGroup, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.Speed != nil {
			// Some devices return -1 if the speed is unknown.
			if *ifaceInfo.Speed >= 0 || !netclassInvalidSpeed {
				speedBytes := int64(*ifaceInfo.Speed * 1000 * 1000 / 8)
				pushMetric(ch, "speed_bytes", speedBytes, ifaceInfo.Name, prometheus.GaugeValue)
			}
		}

		if ifaceInfo.TxQueueLen != nil {
			pushMetric(ch, "transmit_queue_length", *ifaceInfo.TxQueueLen, ifaceInfo.Name, prometheus.GaugeValue)
		}

		if ifaceInfo.Type != nil {
			pushMetric(ch, "protocol_type", *ifaceInfo.Type, ifaceInfo.Name, prometheus.GaugeValue)
		}
	}
	return nil
}

func pushMetric(ch chan<- prometheus.Metric, name string, value int64, ifaceName string, valueType prometheus.ValueType) {
	fieldDesc := prometheus.NewDesc(
		prometheus.BuildFQName("node", "network", name),
		fmt.Sprintf("%s value of /sys/class/net/<iface>.", name),
		[]string{"device"},
		nil,
	)

	ch <- prometheus.MustNewConstMetric(fieldDesc, valueType, float64(value), ifaceName)
}

var (
	netclassIgnoredDevices        = "^$"
	netclassInvalidSpeed          = false
	ignoredNetClassDevicesPattern = regexp.MustCompile(netclassIgnoredDevices)
)
