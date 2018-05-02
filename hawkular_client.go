package main

import (
  "github.com/hawkular/hawkular-client-go/metrics"
  "github.com/prometheus/common/log"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  "os"
  "crypto/tls"
  "io/ioutil"
  "strings"
  "time"
)
// Will be called in main.go. Gathers information from Kubernetes, then gets metrics from Hawkular
func get_metrics() ([]HawkMetric,error) {
  log.Info("HAWKULAR URL: ", os.Getenv("HAWKULAR_URL"))
  log.Info("HAWKULAR TENANT: ", os.Getenv("HAWKULAR_TENANT"))
  // Read the Service Account token to be used to authenticate to Hawkular and to Kubernetes API
  dat, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
  tok := string(dat)
  tok = strings.TrimSpace(tok)
  if err != nil {
    log.Errorf("Failer to get token %v",err)
    return nil,err
  }
  var retmetrics []HawkMetric
  // Set up Configs for Hawkular Client
  tC := &tls.Config{}
  tC.InsecureSkipVerify = true
  p := metrics.Parameters{
    Tenant: os.Getenv("HAWKULAR_TENANT"),
    Url: os.Getenv("HAWKULAR_URL"),
    Token: tok,
    TLSConfig: tC,
  }
  // Get Hawkular Client
  h, err := metrics.NewHawkularClient(p)
  if err != nil {
    log.Errorf("Hawkular Client fail %v",err)
    return nil,err
  }
  log.Info("Got Hawkular Client")
  // Get Kubernetes Cluster Config
  config, err := rest.InClusterConfig()
  if err != nil {
    log.Errorf("Failed getting in Cluster Config %v",err)
    return nil,err
  }
  // Get Kubernetes API Client
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    log.Errorf("Error getting ClientSet %v",err)
    return nil,err
  }
  // Get running pods from Kubernetes API
  pods, err := clientset.CoreV1().Pods(os.Getenv("HAWKULAR_TENANT")).List(metav1.ListOptions{})
  log.Info("Got Pods ",len(pods.Items))
  // Get PVCs from Kubernetes API
  volumes, err := clientset.CoreV1().PersistentVolumeClaims(os.Getenv("HAWKULAR_TENANT")).List(metav1.ListOptions{})
  if err != nil {
     log.Errorf("Error getting pvcs %v",err)
     return nil, err
  }
  log.Info("Got Volumes ",len(volumes.Items))
  // Iterate through pods and volumes to generate tags to narrow down metrics call to Hawkular
  pods_tag :=""
  vol_tag :=""
  for _, pod := range pods.Items {
    pods_tag=pods_tag+pod.Name+"||"
    for _, vol := range pod.Spec.Volumes {
      for _, pvc := range volumes.Items {
        if vol.VolumeSource.PersistentVolumeClaim != nil && pvc.Name == vol.VolumeSource.PersistentVolumeClaim.ClaimName {
          vol_tag=vol_tag+"Volume:"+vol.Name+"||"
        }
      }
    }
  }
  // Get Pod level metrics
  pods_tag=pods_tag[:len(pods_tag)-2]
  tags := make(map[string]string)
  tags["namespace_name"] =  os.Getenv("HAWKULAR_TENANT")
  tags["pod_name"]=pods_tag
  tags["type"]="pod"
  // Get metric Definitions for pods
  mdef, err := h.Definitions(metrics.Filters(metrics.TagsFilter(tags)))
  log.Info("Got metric defs ", len(mdef))
  if len(volumes.Items) > 0 {
    vol_tag=vol_tag[:len(vol_tag)-2]
  }
  tags["resource_id"] = vol_tag
  log.Info(vol_tag)
  // Get metrics Definitions for volumes
  vdef, err := h.Definitions(metrics.Filters(metrics.TagsFilter(tags)))
  log.Info("Got volume defs ", len(vdef))
  // Generate Asynch bool channels
  gmu_done := make(chan bool, 1)
  var gmu []HawkMetric
  gml_done := make(chan bool, 1)
  var gml []HawkMetric

  gcu_done := make(chan bool, 1)
  var gcu []HawkMetric
  gcl_done := make(chan bool, 1)
  var gcl []HawkMetric

  cu_done := make(chan bool, 1)
  var cu []HawkMetric

  nrr_done := make(chan bool, 1)
  var nrr []HawkMetric
  ntr_done := make(chan bool, 1)
  var ntr []HawkMetric

  fa_done := make(chan bool, 1)
  var fa []HawkMetric
  fl_done := make(chan bool, 1)
  var fl []HawkMetric
  fu_done := make(chan bool, 1)
  var fu []HawkMetric
  // Start Asynch calls to get metrics
  go func(){
    gmu = get_metric(h,metrics.Gauge,"memory/usage",mdef)
    gmu_done <- true
  }()
  go func(){
    gml = get_metric(h,metrics.Gauge,"memory/limit",mdef)
    gml_done <- true
  }()
  go func(){
    gcu = get_metric(h,metrics.Gauge,"cpu/usage_rate",mdef)
    gcu_done <- true
  }()
  go func(){
    gcl = get_metric(h,metrics.Gauge,"cpu/limit",mdef)
    gcl_done <- true
  }()
  go func(){
    cu = get_metric(h,metrics.Counter,"uptime",mdef)
    cu_done <- true
  }()
  go func(){
    nrr = get_metric(h,metrics.Gauge,"network/rx_rate",mdef)
    nrr_done <- true
  }()
  go func(){
    ntr = get_metric(h,metrics.Gauge,"network/tx_rate",mdef)
    ntr_done <- true
  }()
  go func(){
    fa = get_metric(h,metrics.Gauge,"filesystem/available",vdef)
    fa_done <- true
  }()
  go func(){
    fl = get_metric(h,metrics.Gauge,"filesystem/limit",vdef)
    fl_done <- true
  }()
  go func(){
    fu = get_metric(h,metrics.Gauge,"filesystem/usage",vdef)
    fu_done <- true
  }()
  // Wait for metrics calls to be finished to append to retmetrics array
  <-gmu_done
  log.Info("Got memory usage:      ", len(gmu),"      ")
  retmetrics = append(retmetrics,gmu...)
  <-gml_done
  log.Info("Got memory limit:      ", len(gml),"      ")
  retmetrics = append(retmetrics,gml...)

  <-gcu_done
  log.Info("Got CPU Usage:         ", len(gcu),"      ")
  retmetrics = append(retmetrics,gcu...)
  <-gcl_done
  log.Info("Got CPU Limit:         ", len(gcl),"      ")
  retmetrics = append(retmetrics,gcl...)

  <-cu_done
  log.Info("Got Uptime:            ", len(cu),"      ")
  retmetrics = append(retmetrics,cu...)

  <-nrr_done
  log.Info("Got Network RX:        ", len(nrr),"      ")
  retmetrics = append(retmetrics,nrr...)
  <-ntr_done
  log.Info("Got Network TX:        ", len(ntr),"      ")
  retmetrics = append(retmetrics,ntr...)

  <-fa_done
  log.Info("Got Filesystem Avail:  ", len(fa),"      ")
  retmetrics = append(retmetrics,fa...)
  <-fl_done
  log.Info("Got Filesystem Limit:  ", len(fl),"      ")
  retmetrics = append(retmetrics,fl...)
  <-fu_done
  log.Info("Got Filesystem Used:  ", len(fu),"      ")
  retmetrics = append(retmetrics,fu...)

  return retmetrics,nil
}

// Get a single metrics type
func get_metric(h *metrics.Client, metric_type metrics.MetricType, metric_name string, metrics_defs []*metrics.MetricDefinition) ([]HawkMetric){
  var retmetrics []HawkMetric
  // Iterate through metric definitions to only get specific metric
  for _, md := range metrics_defs {
    if md.Tags["descriptor_name"] == metric_name {
      dat, err := h.ReadRaw(metric_type,md.ID, metrics.Filters(metrics.LimitFilter(2)))
      if err != nil {
        log.Errorf("Read Raw Failure %v",err)
         return nil
      }
      // Only return fresh data
      if len(dat) > 0  && time.Now().Sub(dat[0].Timestamp).Seconds() < 60 {
        var met HawkMetric
        met.Tags = md.Tags
        met.ID = md.ID
        met.Value = dat[0].Value
        if len(met.Tags["container_name"]) > 0 {
          met.kind = "container"
        } else {
          met.kind = "pod"
        }
        retmetrics = append(retmetrics, met)
      }
    }
  }
  return retmetrics
}
