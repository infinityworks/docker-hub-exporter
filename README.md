# Prometheus Docker Hub Exporter

Exposes metrics of container pulls and stars from the Docker Hub API, to a Prometheus compatible endpoint. The exporter is capable of pulling down stats for individual images, or for orgs or users from DockerHub. This is based on the un-documented V2 Docker Hub API.

## Configuration

The image is setup to take two parameters via command line flags. Below is a list of available flags. You can also find this list by using the `--help` flag.

* `images` Images you wish to monitor: expected format 'user/image1,user/image2'
* `listen-address` Address on which to expose metrics and web interface. (default ":8080")
* `organisations` Organisations/Users you wish to monitor: expected format 'org1,org2'
* `telemetry-path` Path under which to expose metrics. (default "/metrics") 

## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9170:9170 infinityworks/docker-hub-exporter -listen-address=:9170 -images="infinityworks/ranch-eye,infinityworks/prom-conf" -organisations="infinityworks"
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9170:9170 <image-name> -listen-address=:9170 -images="infinityworks/ranch-eye,infinityworks/prom-conf" -organisations="infinityworks"
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/docker-hub-exporter.svg)](http://microbadger.com/images/infinityworks/docker-hub-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/docker-hub-exporter.svg)](http://microbadger.com/images/infinityworks/docker-hub-exporter "Get your own version badge on microbadger.com")
