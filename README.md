# Transmission Exporter for Prometheus

[![Build and Push](https://github.com/khayyamsaleem/transmission-exporter/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/khayyamsaleem/transmission-exporter/actions)

Prometheus exporter for [Transmission](https://transmissionbt.com/) metrics, written in Go.

Fork of [metalmatze/transmission-exporter](https://github.com/metalmatze/transmission-exporter) with tracker labels, modernized logging, additional metrics, bug fixes, and a bundled Grafana dashboard.

### Docker

```sh
docker pull ghcr.io/khayyamsaleem/transmission-exporter:latest
docker run -d -p 19091:19091 ghcr.io/khayyamsaleem/transmission-exporter
```

### Configuration

| ENV Variable | Description | Default |
|---|---|---|
| `WEB_PATH` | Path for metrics | `/metrics` |
| `WEB_ADDR` | Exporter listen address | `:19091` |
| `TRANSMISSION_ADDR` | Transmission RPC address | `http://localhost:9091` |
| `TRANSMISSION_USERNAME` | Transmission username | |
| `TRANSMISSION_PASSWORD` | Transmission password | |

### Docker Compose

```yaml
transmission:
  image: linuxserver/transmission
  restart: always
  ports:
    - "127.0.0.1:9091:9091"
    - "51413:51413"
    - "51413:51413/udp"
transmission-exporter:
  image: ghcr.io/khayyamsaleem/transmission-exporter:latest
  restart: always
  ports:
    - "127.0.0.1:19091:19091"
  environment:
    TRANSMISSION_ADDR: http://transmission:9091
```

### Grafana Dashboard

A dashboard is included in [`dashboards/`](dashboards/). Import `dashboard.json` into Grafana.

### Changes from upstream

- Tracker label on all per-torrent metrics
- Structured logging with `slog`
- Nil pointer crash fixes
- Multi-stage Dockerfile, published to GHCR
- CI via GitHub Actions
- Updated Grafana dashboard

### Development

```sh
cp .env.example .env  # configure Transmission connection
make install           # build and install
```

### Original authors of the Transmission package
Tobias Blom (https://github.com/tubbebubbe/transmission)
Long Nguyen (https://github.com/longnguyen11288/go-transmission)
