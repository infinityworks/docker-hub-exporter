FROM python:3.5-alpine

RUN pip install prometheus_client requests

ENV BIND_PORT 1234
ENV IMAGES "prom/prometheus, prom/node-exporter"

ADD . /usr/src/app
WORKDIR /usr/src/app

CMD ["python", "hub_exporter.py"]
