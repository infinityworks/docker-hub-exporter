# Prometheus Docker Hub Exporter

Exposes metrics of container pulls and stars from the Docker Hub API, to a Prometheus compatible endpoint. 

## Configuration

The image is setup to take two parameters from environment variables:
* `BIND_PORT` The port you wish to run the container on, defaults to 9170
* `ORGS` The docker hub organizations you wish to monitor, expected in the format "org1, org2"
* `IMAGES` The images you wish to monitor, expected in the format "user/image1, user/image2". Can be across different dockerhub users.

## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9170:9170 -e IMAGES="infinityworks/ranch-eye, infinityworks/prom-conf" -e ORGS="infinityworks" infinityworks/docker-hub-exporter
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9170:9170 -e IMAGES="infinityworks/ranch-eye, infinityworks/prom-conf" <image-name>
```

## Docker compose

```
docker-hub-exporter:
    tty: true
    stdin_open: true
    expose:
      - 9170:9170
    image: infinityworks/docker-hub-exporter
```

## Metrics

Metrics will be made available on port 9170 by default

```
# HELP docker_hub_pulls_total counter of docker_pulls from the public API
# TYPE docker_hub_pulls_total counter
docker_hub_pulls_total{image="prometheus-rancher-exporter",user="infinityworks"} 163994.0
docker_hub_pulls_total{image="docker-hub-exporter",user="infinityworks"} 40.0
# HELP docker_hub_stars gauge of docker_stars from the public API
# TYPE docker_hub_stars gauge
docker_hub_stars{image="prometheus-rancher-exporter",user="infinityworks"} 3.0
docker_hub_stars{image="docker-hub-exporter",user="infinityworks"} 1.0
# HELP docker_hub_is_automated gauge of is_automated from the public API
# TYPE docker_hub_is_automated gauge
docker_hub_is_automated{image="prometheus-rancher-exporter",user="infinityworks"} 1.0
docker_hub_is_automated{image="docker-hub-exporter",user="infinityworks"} 1.0
# HELP docker_hub_last_updated_seconds Unix timestamp of last_updated from the public API
# TYPE docker_hub_last_updated_seconds gauge
docker_hub_last_updated_seconds{image="prometheus-rancher-exporter",user="infinityworks"} 1472731040.0
docker_hub_last_updated_seconds{image="docker-hub-exporter",user="infinityworks"} 1473664481.0
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/docker-hub-exporter.svg)](http://microbadger.com/images/infinityworks/docker-hub-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/docker-hub-exporter.svg)](http://microbadger.com/images/infinityworks/docker-hub-exporter "Get your own version badge on microbadger.com")
