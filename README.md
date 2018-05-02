# hawkular_exporter
Hawkular exporter for Prometheus.

The purpose of this exporter was to get the metrics of hawkular that supports multi-tenancy into Prometheus so metrics can be used for creating and managing alerts by either Grafana or AlertManager.

NOTE:

This image is meant to be run in Openshift/Kubernetes. As such, it expects a service account token with View Access to the Namespace to be mounted to the Pod. The token file should be located at: "/var/run/secrets/kubernetes.io/serviceaccount/token"

## Installation

## Usage of `hawkular_exporter`
Help is displayed with `-h`.

| Option                   | Default             | Description
| ------------------------ | ------------------- | -----------------
| -help                    | -                   | Displays usage.
| -listen-address          | `:9189`             | The address to listen on for HTTP requests.
| -log.format              | `logger:stderr`     | Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
| -log.level               | `info`              | Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
| -metrics-path            | `/metrics`          | URL Endpoint for metrics
| -version                 | -                   | Prints version information


## Env Variables to Set:
Using the Hawkular Go Client to call metrics REST API on authenticated hawkular will require these environment variables to be set:

* HAWKULAR_URL: The Hawkular Server (eg. https://metrics.domain.org)

* HAWKULAR_TENANT: The Namespace in which the exporter is running.

### Metrics

| Name                                 | Description                              |
| ------------------------------------ | -----------------------------------------|
| hawkular_up                          | Was the last call to Hawkular successful |
| hawkular_mem_usage_bytes             | Memory usage of each pod in Namespace    |
| hawkular_mem_limit_bytes             | Memory limit of each pod in Namespace    |
| hawkular_cpu_usage_millicores        | CPU usage of each pod in Namespace       |
| hawkular_cpu_limit_millicores        | CPU limit of each pod in Namespace       |
| hawkular_network_tx_bytes_seconds    | Network TX of each pod in Namespace      |
| hawkular_network_rx_bytes_seconds    | Network RX of each pod in Namespace      |
| hawkular_filesystem_usage_bytes      | Filesystem usage of each Persistent Volume attached to Pods in the Namespace      |
| hawkular_filesystem_available_bytes  | Filesystem available of each Persistent Volume attached to Pods in the Namespace  |
| hawkular_filesystem_limit_bytes      | Filesystem limit of each Persistent Volume attached to Pods in the Namespace      |
| hawkular_uptime_milliseconds         | Uptime of each Pod in Namespace |

### Building Locally

Make sure you have make, go, and docker installed.

To build the Go Binary without dependencies encapsulated:

```
make build_fast
```

To build the full binary with no dependencies:
```
make build
```

To build the Docker Image:
```
make docker
```
