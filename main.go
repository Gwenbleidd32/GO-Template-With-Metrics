// This is a Simple HTTP Go Server meant to test metrics usage with Prometheus. 
// I wrote this to use with attatched sidecars in cloud run to test the native 
// PROM-QL queries  within cloud monitoring
// 

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)
// Prometheus tools

// Define custom metrics for load testing
var (
	// Measure the duration of HTTP requests
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Count the total number of HTTP requests
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)

	// Count the number of request errors
	requestErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of failed HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)
)

func init() {
	// Register custom metrics with Prometheus
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestErrors)
}


// Collector that contains the descriptors for the metrics from the app.
// Foo is a gauge with no labels. Bar is a counter with no labels.
type fooBarCollector struct {
	fooMetric *prometheus.Desc
	barMetric *prometheus.Desc
}

func newFooBarCollector() *fooBarCollector {
	return &fooBarCollector{
		fooMetric: prometheus.NewDesc("foo_metric",
			"A foo event has occurred",
			nil, nil,
		),
		barMetric: prometheus.NewDesc("bar_metric",
			"A bar event has occured",
			nil, nil,
		),
	}
}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *fooBarCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.fooMetric
	ch <- collector.barMetric
}

// Collect implements required collect function for all prometheus collectors
func (collector *fooBarCollector) Collect(ch chan<- prometheus.Metric) {
	m1 := prometheus.MustNewConstMetric(collector.fooMetric, prometheus.GaugeValue, float64(time.Now().Unix()))
	m2 := prometheus.MustNewConstMetric(collector.barMetric, prometheus.CounterValue, float64(time.Now().Unix()))
	ch <- m1
	ch <- m2
}

func entrypointHandler(w http.ResponseWriter, r *http.Request) {
	// Start Measuring the time before hhandling request
	start := time.Now()

	// Increment the request count
	requestCount.WithLabelValues(r.Method, r.URL.Path).Inc()
	// Increment request duration
	requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())


	// HTML response
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    fmt.Fprintln(w, `
        <!doctype html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1">
            <title>App-1</title>
            <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
            <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Raleway">
            <style>
                body,h1 {font-family: "Raleway", sans-serif}
                body, html {height: 100%}
                .bgimg {
                  background-image: url('https://storage.googleapis.com/a-dream/Houses.jpg');
                  min-height: 100%;
                  background-position: center;
                  background-size: cover;
                }

                .w3-display-middle {
                  background-image: url('https://storage.googleapis.com/a-dream/blue.png');
                  background-size: cover;
                  padding: 200px;
                  border-radius: 25px;
                  text-align: center;
                }

                button { 
                  font-size: 2em; 
                  padding: 10px 20px; 
                  background: transparent;
                  color: white;
                  border: 2px solid white;
                  cursor: pointer;
                }

                button:hover {
                  background: rgba(255, 255, 255, 0.2);
                }
            </style>
        </head>
        <body>
		<div class="bgimg w3-display-container w3-animate-opacity w3-text-white">
		<div class="w3-display-middle">
			<h1>GERMANY</h1>
			<button onclick="window.location.href='https://github.com/Gwenbleidd32/GO-Template-With-Metrics/tree/main'">GITHUB</button>
		</div>
		</div>

        </body>
        </html>
    `)
}


func main() {
	foo := newFooBarCollector()
	prometheus.MustRegister(foo)

	entrypointMux := http.NewServeMux()
	entrypointMux.HandleFunc("/", entrypointHandler)
	entrypointMux.HandleFunc("/startup", entrypointHandler)
	entrypointMux.HandleFunc("/liveness", entrypointHandler)

	promMux := http.NewServeMux()
	promMux.Handle("/metrics", promhttp.Handler())

	go func() {
		http.ListenAndServe(":8000", entrypointMux)
	}()

	http.ListenAndServe(":8080", promMux)
}
