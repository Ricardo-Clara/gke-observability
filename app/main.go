package main

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequests = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    },
    []string{"path", "status"},
)

func init() {
    prometheus.MustRegister(httpRequests)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    httpRequests.WithLabelValues("/health", "200").Inc()
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
}

func main() {
    http.HandleFunc("/health", healthHandler)
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}