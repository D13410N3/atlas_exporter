# Atlas Probe Metrics Exporter

## A simple Prometheus exporter written in Go that fetches metrics from Atlas probes via the RIPE Atlas API. The exporter serves metrics at individual endpoints for each Atlas probe ID.

## Features
* Fetches data from https://atlas.ripe.net/api/v2/probes/{id}/
* Handles requests in the format /metrics/{id}
* Serves metrics with a custom Prometheus registry for each request


## Metrics Exposed
* `atlas_status`: Atlas Probe Status ID
* `atlas_status_since`: Atlas Probe Status Since (timestamp)
* `atlas_total_uptime`: Atlas Probe Total Uptime (in seconds)
* `atlas_first_connected` Atlas Probe was reg at the first time
* `atlas_last_connected` Atlas Probe was reg at the last time
* `atlas_info`: Atlas Probe Info with labels address_v4 and address_v6

## Prerequisites
* Go (tested with version 1.20.5)
* Prometheus Go client library
* Gorilla Mux router

## Installation

### Building from Source
* Clone this repository:
`git clone https://github.com/D13410N3/atlas_exporter.git`
* Navigate to the directory and build:
```
cd atlas_probe_exporter
go build
```

## Running
* Optionally, set the LISTEN_ADDR environment variable to specify the listening address and port. The default is 0.0.0.0:9207.
```
export LISTEN_ADDR=0.0.0.0:9207

./atlas_probe_exporter
```

* Open a browser or use curl to fetch metrics for a specific Atlas probe ID:

`curl http://localhost:9207/metrics/123456`
