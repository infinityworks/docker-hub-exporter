from prometheus_client import start_http_server
from prometheus_client.core import CounterMetricFamily, GaugeMetricFamily, REGISTRY

import json
import requests
import sys
import time
import os


class HubCollector(object):

  def collect(self):
      images = os.getenv('IMAGES', default="thebsdbox/ovcli, rucknar/prom-conf").replace(' ','').split(",")
      print("Starting exporter")
      self._pull_metrics = CounterMetricFamily('docker_hub_pulls_total', 'counter of docker_pulls from the public API', labels=["image", "user"])
      self._star_metrics = GaugeMetricFamily('docker_hub_stars', 'gauge of docker_stars from the public API', labels=["image", "user"])
      
      for image in images:
          print("Getting JSON for " + image)
          self._get_json(image)
          print("Getting Metrics for " + image)
          self._get_metrics()
          print ("Metrics Updated for " + image)

      yield self._pull_metrics
      yield self._star_metrics

  def _get_json(self, image):
      print("Getting JSON Payload for " + image)
      image_url = 'https://hub.docker.com/v2/repositories/{0}'.format(image)
      print(image_url)
      response = requests.get(image_url)
      self._response_json = json.loads(response.content.decode('UTF-8'))


  def _get_metrics(self):
      image_name = self._response_json['name']
      user_name = self._response_json['user']
      self._pull_metrics.add_metric([image_name, user_name], value=self._response_json['pull_count'])
      self._star_metrics.add_metric([image_name, user_name], value=self._response_json['star_count'])


if __name__ == '__main__':
  start_http_server(int(os.getenv('BIND_PORT')))
  REGISTRY.register(HubCollector())

  while True: time.sleep(1)
