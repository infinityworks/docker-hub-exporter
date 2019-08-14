package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"

	"fmt"

	exporter "github.com/infinityworks/docker-hub-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		listenAddress     = flag.String("listen-address", ":9170", "Address on which to expose metrics and web interface.")
		metricsPath       = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
		flagOrgs          = flag.String("organisations", "", "Organisations/Users you wish to monitor: expected format 'org1,org2'")
		flagImages        = flag.String("images", "", "Images you wish to monitor: expected format 'user/image1,user/image2'")
		connectionRetries = flag.Int("connection-retries", 3, "Connection retries until failure is raised. ")
		connectionTimeout = flag.Int("connection-timeout", 5, "Connection timeout in seconds. ")
	)

	var organisations []string
	var images []string
	
	envBind := os.Getenv("BIND_PORT")
	envOrgs := os.Getenv("ORGS")
	envImages := os.Getenv("IMAGES")
	envConnectionRetries := os.Getenv("CONNECTION_RETRIES")
	envConnectionTimeout := os.Getenv("CONNECTION_TIMEOUT")

	flag.Parse()

	if *flagOrgs == "" && envOrgs == "" && *flagImages == "" && envImages == "" {
		log.Fatal("No organisations or images provided")
	}

	if envBind != "" {
		listenAddress = &envBind
	}

	if envConnectionRetries != "" {
		var err error
		*connectionRetries, err = strconv.Atoi(envConnectionRetries)

		if err != nil {
			log.Fatal("Invalid value for connection-retries")
		}
	}

	if envConnectionTimeout != "" {
		var err error
		*connectionTimeout, err = strconv.Atoi(envConnectionTimeout)

		if err != nil {
			log.Fatal("Invalid value for connection-timeout")
		}
	}

	organisations = append(organisations, strings.Split(*flagOrgs, ",")...)
	images = append(images, strings.Split(*flagImages, ",")...)

	if envOrgs != "" {
		organisations = append(organisations, strings.Split(envOrgs, ",")...)
	}

	if envImages != "" {
		images = append(images, strings.Split(envImages, ",")...)
	}

	if strings.HasPrefix(*listenAddress, ":") != true {
		*listenAddress = fmt.Sprintf(":%s", *listenAddress)
	}

	log.Println("Starting Docker Hub Exporter")
	log.Printf("Listening on: %s", *listenAddress)

	e := exporter.New(
		organisations,
		images,
		*connectionRetries,
		exporter.WithLogger(log.New(os.Stdout, "docker_hub_exporter: ", log.LstdFlags)),
		exporter.WithTimeout(time.Second*time.Duration(*connectionTimeout)),
	)

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(*e)

	// Setup HTTP handler
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		                <head><title>Docker Hub Exporter</title></head>
		                <body>
		                   <h1>Docker Hub Prometheus Metrics Exporter</h1>
				   <p>For more information, visit <a href='https://github.com/infinityworks/docker-hub-exporter'>GitHub</a></p>
		                   <p><a href='` + *metricsPath + `'>Metrics</a></p>
		                   </body>
		                </html>
		              `))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
