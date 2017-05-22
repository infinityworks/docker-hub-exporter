package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/infinityworksltd/docker-hub-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	var (
		listenAddress = flag.String("listen-address", ":9171", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
		organisations = flag.String("organisations", "", "Organisations/Users you wish to monitor: expected format 'org1,org2'")
		images        = flag.String("images", "", "Images you wish to monitor: expected format 'user/image1,user/image2'")
	)

	flag.Parse()

	if *organisations == "" && *images == "" {
		log.Fatal("No organisations or images provided")
	}

	log.Println("Starting Docker Hub Exporter")
	log.Printf("Listening on: %s", *listenAddress)

	exporter := exporter.New(
		strings.Split(*organisations, ","),
		strings.Split(*images, ","),
		exporter.WithLogger(log.New(os.Stdout, "docker_hub_exporter: ", log.LstdFlags)),
		exporter.WithTimeout(time.Second*1),
	)

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(*exporter)

	// Setup HTTP handler
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		                <head><title>Docker Hub Exporter</title></head>
		                <body>
		                   <h1>Docker Hub Prometheus Metrics Exporter</h1>
						   <p>For more information, visit <a href=https://github.com/infinityworksltd/docker-hub-exporter>GitHub</a></p>
		                   <p><a href='` + *metricsPath + `'>Metrics</a></p>
		                   </body>
		                </html>
		              `))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
