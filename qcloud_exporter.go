package main

import (
	"net/http"

	"github.com/tencentyun/tencentcloud-exporter/collector"
	"github.com/tencentyun/tencentcloud-exporter/config"
	"github.com/tencentyun/tencentcloud-exporter/instances"
	"github.com/tencentyun/tencentcloud-exporter/monitor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {

	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").
				Default("localhost:8001").String()

		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").
				Default("/metrics").String()

		configFile = kingpin.Flag("config.file", "Tencent qcloud exporter configuration file.").
				Default("qcloud.yml").String()
	)
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("tencent_qcloud_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting tencent_qcloud_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	tencentConfig := config.NewConfig()
	if err := tencentConfig.LoadFile(*configFile); err != nil {
		log.Fatal(err.Error())
	}

	credentialConfig := (*tencentConfig).Credential
	metricsConfig := (*tencentConfig).Metrics

	if err := monitor.InitClient(credentialConfig, (*tencentConfig).RateLimit); err != nil {
		log.Fatal(err.Error())
	}

	if err := instances.InitClient(credentialConfig); err != nil {
		log.Fatal(err.Error())
	}

	tencentCollector, err := collector.NewCollector(metricsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(tencentCollector)

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
