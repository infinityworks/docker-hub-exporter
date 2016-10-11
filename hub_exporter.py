from prometheus_client import start_http_server
from prometheus_client.core import CounterMetricFamily, GaugeMetricFamily, REGISTRY

import json, requests, sys, time, os, ast, signal, logging, datetime, calendar

class GitHubCollector(object):

  def collect(self):

    metrics = {'stars': ['star_count', 'GaugeMetricFamily'],
               'is_automated': ['is_automated', 'GaugeMetricFamily'],
               'pulls_total': ['pull_count', 'CounterMetricFamily'],
               'last_updated': ['last_updated', 'GaugeMetricFamily']
               }

    METRIC_PREFIX = 'docker_hub_image'
    LABELS = ['image', 'user']
    data = {}

    # Setup metric counters from prometheus_client.core
    for metric, field in metrics.items():
      if field[1] == "GaugeMetricFamily":
        data[metric] = GaugeMetricFamily('%s_%s' % (METRIC_PREFIX, metric), '%s' % metric, value=None, labels=LABELS)
      elif field[1] == "CounterMetricFamily":
        data[metric] = CounterMetricFamily('%s_%s' % (METRIC_PREFIX, metric), '%s' % metric, value=None, labels=LABELS)

    # loop through specified images and organizations and collect metrics
    if os.getenv('IMAGES'):
      images = os.getenv('IMAGES').replace(' ','').split(",")
      self._image_urls = []
      for image in images:
        self._image_urls.extend('https://hub.docker.com/v2/repositories/{0}'.format(image).split(","))
      self._collect_image_metrics(data, metrics)

    if os.getenv('ORGS'):
      orgs = os.getenv('ORGS').replace(' ','').split(",")
      self._org_urls = []
      for org in orgs:
        self._org_urls.extend('https://hub.docker.com/v2/repositories/{0}'.format(org).split(","))
      self._collect_org_metrics(data, metrics)

    # Yield all metrics returned
    for metric in metrics:
      yield data[metric]

  def _collect_image_metrics(self, data, metrics):
    for image in self._image_urls:
      print('Collecting metrics for image:  ' + image)
      response_json = self._get_json(image)
      self._convert_to_timestamps(response_json)
      self._add_metrics(data, metrics, response_json)

  def _collect_org_metrics(self, data, metrics):
    for org_url in self._org_urls:
      full_content = []
      page_ref = org_url
      while True:
        page_content = self._get_json(page_ref)
        full_content = full_content + page_content['results']
        if page_content['next']:
          page_ref = page_content['next']
        else:
          break
      for image in full_content:
        updated_image = self._convert_to_timestamps(image)
        self._add_metrics(data, metrics, updated_image)
        print("Adding metrics for" + updated_image['name'])

  def _get_json(self, url):
    response = requests.get(url)
    response_json = json.loads(response.content.decode('UTF-8'))
    return response_json

  def _convert_to_timestamps(self, response_json):
    last_updated_pre_conversion = datetime.datetime.strptime(response_json['last_updated'], "%Y-%m-%dT%H:%M:%S.%fZ")
    response_json['last_updated'] = float(calendar.timegm(last_updated_pre_conversion.utctimetuple()))
    return response_json

  def _add_metrics(self, data, metrics, response_json):
    for metric, field in metrics.items():
      data[metric].add_metric([response_json['name'], response_json['user']], value=response_json[field[0]])

def sigterm_handler(_signo, _stack_frame):
  sys.exit(0)

if __name__ == '__main__':
  # Ensure we have something to export
  if not (os.getenv('IMAGES') or os.getenv('ORGS')):
    print("No Images or organizations specified, exiting")
    exit(1)
  start_http_server(int(os.getenv('BIND_PORT')))
  REGISTRY.register(GitHubCollector())
  
  signal.signal(signal.SIGTERM, sigterm_handler)
  while True: time.sleep(1)