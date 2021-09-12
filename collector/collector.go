package collector

import (
	"exporter/parser"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)

	collectorManager *CollectorManager
)

func init() {
	linuxParser, err := parser.NewLinuxParser()
	if err != nil {
		panic(err)
	}

	collectorManager = &CollectorManager{
		collectors: make(map[string]Collector),
		Parser:     linuxParser,
	}
}

func GetCollectorManager() *CollectorManager {
	return collectorManager
}

func registerCollector(collector Collector) error {
	return collectorManager.registerCollector(collector)
}

func UnRegisterCollector(name string) error {
	return collectorManager.unRegisterCollector(name)
}

type Collector interface {
	GetName() string
	Describe(chan<- *prometheus.Desc) error
	Collect(parser.Parser, chan<- prometheus.Metric) error
}

type CollectorManager struct {
	collectors map[string]Collector
	parser.Parser
}

func (manager *CollectorManager) registerCollector(collector Collector) error {
	if _, exist := manager.collectors[collector.GetName()]; exist {
		return fmt.Errorf("collector %s is already exist", collector.GetName())
	}
	manager.collectors[collector.GetName()] = collector
	return nil
}

func (manager *CollectorManager) unRegisterCollector(name string) error {
	if _, exist := manager.collectors[name]; !exist {
		return fmt.Errorf("collector %s is not register", name)
	}

	delete(manager.collectors, name)
	return nil
}

func (manager *CollectorManager) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range manager.collectors {
		if err := collector.Describe(ch); err != nil {
			// manager.Printf("[ERROR] collector %s describe failed, error: %s", collector.GetName(), err.Error())
			log.Println(fmt.Sprintf("[ERROR] collector %s describe failed, error: %s", collector.GetName(), err.Error()))
			continue
		}
	}
}

func (manager *CollectorManager) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(manager.collectors))

	for _, collector := range manager.collectors {
		go func(collector Collector, ch chan<- prometheus.Metric) {
			defer wg.Done()

			now := time.Now()
			var success float64
			name := collector.GetName()
			if err := collector.Collect(manager.Parser, ch); err != nil {
				// manager.Printf("[ERROR] collector %s collect failed, error: %s", name, err.Error())
				log.Println(fmt.Sprintf("[ERROR] collector %s collect failed, error: %s", name, err.Error()))
			} else {
				success = 1
			}
			duration := time.Since(now)

			ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
		}(collector, ch)
	}
	wg.Wait()
}
