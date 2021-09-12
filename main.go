package main

import (
	"exporter/collector"
	"exporter/parser"
	"exporter/registry"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricPath      string
	port            int
	exporterTags    string
	registryAddress string
)

func init() {
	flag.StringVar(&metricPath, "metric_path", "metrics", "metric request path")
	flag.IntVar(&port, "port", 9291, "metric http port")

	flag.StringVar(&exporterTags, "exporter_tags", "metrics", "exporter tags")
	flag.StringVar(&registryAddress, "registrt_address", "localhost:8500", "exporter tags")

	flag.Parse()
}

func main() {
	ips, err := parser.GetIPs()
	if err != nil {
		panic(err)
	}

	tags := strings.Split(exporterTags, ",")
	for i := 0; i < 4; i++ {
		time.Sleep(time.Second * 3)
		reg, err := registry.NewConsulRegistry(registry.RegistryConsulConfig{
			Address: registryAddress,
			Schema:  "http",
			Kind:    "node-exporter",
			ID:      ips[0],
			Name:    ips[0],
			Tags:    tags,
		})
		if err != nil {
			log.Println(err)
			continue
		}

		if err := reg.ServiceRegister(ips[0], port, ""); err != nil {
			log.Println(err)
			continue
		}
		defer reg.ServiceUnRegister()
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.GetCollectorManager())

	http.Handle(fmt.Sprintf("/%s", metricPath), promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("node-exporter"))
		rw.WriteHeader(http.StatusOK)
	})

	log.Println("service start at", fmt.Sprintf("0.0.0.0:%d", port))

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil); err != nil {
		panic(err)
	}
}
