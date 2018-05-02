package main

type HawkMetric struct {
  kind      string            `json:"type"`
  ID        string            `json:"id"`
  Tags      map[string]string `json:"tags,omitempty"`
  Value     interface{}       `json:"value"`
}


