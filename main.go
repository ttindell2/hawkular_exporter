// hawkular exporter, exports metrics using Hawkular Go Client
package main

import (
  "flag"
  "fmt"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/prometheus/common/log"
  "github.com/prometheus/common/version"
  "net/http"
  "os"
  "strings"
)

const (
  namespace = "hawkular"
)

var (
  up = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "", "up"),
	"Was the last query of hawkular successful",
	nil, nil,
  )
  mem_usage = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "mem_usage_bytes"),
        "Memory Usage of the pod",
        []string{"kind","pod_name","container_name","namespace"}, nil,
  )
  mem_limit = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "mem_limit_bytes"),
        "Memory Limit of the pod",
        []string{"kind","pod_name","container_name","namespace"}, nil,
  )
  cpu_usage = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "", "cpu_usage_millicores"),
	"CPU Usage of the pod",
	[]string{"kind","pod_name","container_name","namespace"}, nil,
  )
  cpu_limit = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "cpu_limit_millicores"),
        "CPU Limit of the pod",
        []string{"kind","pod_name","container_name","namespace"}, nil,
  )
  uptime = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "uptime_milliseconds"),
        "Uptime of the pod",
        []string{"kind","pod_name","container_name","namespace"}, nil,
  )
  network_tx_rate = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "network_tx_bytes_seconds"),
        "Network TX per sec of the pod",
        []string{"pod_name","namespace"}, nil,
  )
  network_rx_rate = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "network_rx_bytes_seconds"),
        "Network RX per sec of the pod",
        []string{"pod_name","namespace"}, nil,
  )
  filesystem_available = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "filesystem_available_bytes"),
        "Filesystem available of the volume of the pod",
        []string{"volume_name","pod_name","namespace"}, nil,
  )
  filesystem_limit = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "filesystem_limit_bytes"),
        "Filesystem limit of the volume of the pod",
        []string{"volume_name","pod_name","namespace"}, nil,
  )
  filesystem_usage = prometheus.NewDesc(
        prometheus.BuildFQName(namespace, "", "filesystem_usage_bytes"),
        "Filesystem limit of the volume of the pod",
        []string{"volume_name","pod_name","namespace"}, nil,
  )



)

// Exporter holds name, path and volumes to be monitored
type Exporter struct {
  hostname string
}

// Describe all the metrics exported by Hawkular exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
  ch <- up
  ch <- cpu_usage
  ch <- cpu_limit      
  ch <- mem_usage
  ch <- mem_limit
  ch <- uptime
  ch <- network_tx_rate
  ch <- network_rx_rate
  ch <- filesystem_available
  ch <- filesystem_limit
  ch <- filesystem_usage
}

// Collect collects all the metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
  metrics,err := get_metrics()
  if err != nil {
    log.Info("couldn't get metrics info: %v", err)
    ch <- prometheus.MustNewConstMetric(
      up, prometheus.GaugeValue, 0.0,
    )
  } else {
    ch <- prometheus.MustNewConstMetric(
      up, prometheus.GaugeValue, 1.0,
    )
  }
  for _, metric := range metrics {
    switch metric.Tags["descriptor_name"] {
      case "memory/usage":
          ch <- prometheus.MustNewConstMetric(
             mem_usage, prometheus.GaugeValue, metric.Value.(float64),metric.kind,metric.Tags["pod_name"],metric.Tags["container_name"], metric.Tags["namespace_name"],
           )
      case "memory/limit":
          ch <- prometheus.MustNewConstMetric(
             mem_limit, prometheus.GaugeValue, metric.Value.(float64),metric.kind,metric.Tags["pod_name"],metric.Tags["container_name"], metric.Tags["namespace_name"],
           )
      case "cpu/usage_rate":
           ch <- prometheus.MustNewConstMetric(
             cpu_usage, prometheus.GaugeValue, metric.Value.(float64),metric.kind,metric.Tags["pod_name"],metric.Tags["container_name"], metric.Tags["namespace_name"],
           )
      case "cpu/limit":
        ch <- prometheus.MustNewConstMetric(
           cpu_limit, prometheus.GaugeValue, metric.Value.(float64),metric.kind,metric.Tags["pod_name"],metric.Tags["container_name"], metric.Tags["namespace_name"],
         )
      case "uptime":
        ch <- prometheus.MustNewConstMetric(
           uptime, prometheus.CounterValue, metric.Value.(float64),metric.kind,metric.Tags["pod_name"],metric.Tags["container_name"], metric.Tags["namespace_name"],
         )
      case "network/tx_rate":
        ch <- prometheus.MustNewConstMetric(
           network_tx_rate, prometheus.GaugeValue, metric.Value.(float64),metric.Tags["pod_name"],metric.Tags["namespace_name"],
         )
      case "network/rx_rate":
        ch <- prometheus.MustNewConstMetric(
           network_rx_rate, prometheus.GaugeValue, metric.Value.(float64),metric.Tags["pod_name"],metric.Tags["container_name"],
         )
      case "filesystem/available":
        if strings.Contains(metric.ID, "Volume:") {
          var vol_name = strings.Split(metric.ID,"Volume:")[1]
          ch <- prometheus.MustNewConstMetric(
             filesystem_available, prometheus.GaugeValue, metric.Value.(float64),vol_name,metric.Tags["pod_name"],metric.Tags["namespace_name"],
           )
        }
      case "filesystem/limit":
        if strings.Contains(metric.ID,"Volume:") {
          var vol_name = strings.Split(metric.ID,"Volume:")[1]
          ch <- prometheus.MustNewConstMetric(
             filesystem_limit, prometheus.GaugeValue, metric.Value.(float64),vol_name,metric.Tags["pod_name"],metric.Tags["namespace_name"],
          )
        }
      case "filesystem/usage":
        if strings.Contains(metric.ID,"Volume:") {
          var vol_name = strings.Split(metric.ID,"Volume:")[1]
          ch <- prometheus.MustNewConstMetric(
             filesystem_usage, prometheus.GaugeValue, metric.Value.(float64),vol_name,metric.Tags["pod_name"],metric.Tags["namespace_name"],
          )
        }
      default:
        log.Debug(metric.Tags["descriptor_name"] + " not yet supported")
    }
  }
   log.Info("<--------------- Finished collecting metrics --------------->")
}

// NewExporter initialises exporter
func NewExporter(hostname string) (*Exporter, error) {
  return &Exporter{
    hostname: hostname,
  }, nil
}

func versionInfo() {
  fmt.Println(version.Print("hawkular_exporter"))
  os.Exit(0)
}

func init() {
  prometheus.MustRegister(version.NewCollector("hawkular_exporter"))
}

func main() {
  // commandline arguments
  var (
    metricPath    = flag.String("metrics-path", "/metrics", "URL Endpoint for metrics")
      listenAddress = flag.String("listen-address", ":9189", "The address to listen on for HTTP requests.")
      showVersion   = flag.Bool("version", false, "Prints version information")
  )
  flag.Parse()
  if *showVersion {
    versionInfo()
  }
  hostname, err := os.Hostname()
  if err != nil {
    log.Fatalf("While trying to get Hostname error happened: %v", err)
  }
  exporter, err := NewExporter(hostname)
  if err != nil {
    log.Errorf("Creating new Exporter went wrong, ... \n%v", err)
  }
  prometheus.MustRegister(exporter)
  log.Info("Hawkular Metrics Exporter v", version.Version)
  http.Handle("/metrics", promhttp.Handler())
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`<html>
      <head><title>Hawkular Exporter v` + version.Version + `</title></head>
      <body>
      <h1>Hawkular Exporter v` + version.Version + `</h1>
      <p><a href='` + *metricPath + `'>Metrics</a></p>
      </body>
      </html>
    `))
  })
  log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
