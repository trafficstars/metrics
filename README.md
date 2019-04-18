[![Build Status](https://travis-ci.org/trafficstars/metrics.svg?branch=master)](https://travis-ci.org/trafficstars/metrics)
[![go report](https://goreportcard.com/badge/github.com/trafficstars/metrics)](https://goreportcard.com/report/github.com/trafficstars/metrics)
[![GoDoc](https://godoc.org/github.com/trafficstars/metrics?status.svg)](https://godoc.org/github.com/trafficstars/metrics)

Description
===========

This is a implementation of high performance handy metrics for Golang which could be
used for prometheus (passive export) and/or StatsD (active export). But the primary method is
passive export (a special page where somebody get fetch all the metrics).


How to use
==========

Count number of requests (and request rate by measuring the rate of the count):
```go
metrics.Count(`requests`, metrics.Tags{
    `method`: request.Method,
}).Increment()
```

Measure latency:
```go
startTime := time.Now()

[... do your routines here ...]

metrics.TimingBuffered(`latency`, nil).ConsiderValue(time.Since(startTime))
```

Export the metrics for prometheus:
```go
import "github.com/trafficstars/statuspage"

func sendMetrics(w http.ResponseWriter, r *http.Request) {
    statuspage.WriteMetricsPrometheus(w)
}

func main() {
[...]
    http.HandleFunc("/metrics.prometheus", sendMetrics)
[...]
}
```

Export the metrics to StatsD
```go

import (
	"github.com/trafficstars/metrics"
)

func newStatsdSender(address string) (*statsdSender, error) {
[... init ...]
}

func (sender *statsdSender) SendInt64(metric metrics.Metric, key string, int64) error {
[... send the metric to statsd ...]
}

func main() {
[...]
    metricsSender, err := newStatsdSender(`localhost:8125`)
    if err != nil {
		log.Fatal(err)
    }
    metrics.SetDefaultSender(metricsSender)
[...]
}
```

Hello world
-----------

```go
package main

import (
        "fmt"
        "math/rand"
        "net/http"
        "time"

        "github.com/trafficstars/metrics"
        "github.com/trafficstars/statuspage"
)

func hello(w http.ResponseWriter, r *http.Request) {
    answerInt := rand.Intn(10)

    startTime := time.Now()

    // just a metric
    tags := metrics.Tags{`answer_int`: answerInt}
    metrics.Count(`hello`, tags).Increment()

    time.Sleep(time.Millisecond)
    fmt.Fprintf(w, "Hello world! The answerInt == %v\n", answerInt)

    // just a one more metric
    tags["endpoint"] = "hello"
    metrics.TimingBuffered(`latency`, tags).ConsiderValue(time.Since(startTime))
}

func sendMetrics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

    statuspage.WriteMetricsPrometheus(w)

    metrics.TimingBuffered(`latency`, metrics.Tags{
		"endpoint": "sendMetrics",
    }).ConsiderValue(time.Since(startTime))
}

func main() {
    http.HandleFunc("/", hello)
    http.HandleFunc("/metrics.prometheus", sendMetrics) // here we export metrics for prometheus
    http.ListenAndServe(":8000", nil)
}
```

Framework "echo"
----------------

The same as above, but just use our handler:
```go
// import "github.com/trafficstars/statuspage/handler/echostatuspage"

r := echo.New()
r.GET("/status.prometheus", echostatuspage.StatusPrometheus)
```

Aggregative metrics
===================

Aggregative metrics are similar to prometheus' [summary](https://prometheus.io/docs/concepts/metric_types/#summary).
There're available three methods of summarizing/aggregation of observed values:
* Simple.
* Flow.
* Buffered.

There's two types of Aggregative metrics:
* Timing (receives `time.Duration` as the argument to method `ConsiderValue`).
* Gauge (receives `float64` as the argument to method `ConsiderValue`).

`ConsiderValue` is analog of prometheus' [Observe](https://godoc.org/github.com/prometheus/client_golang/prometheus#Summary)

So there're available next aggregative metrics:
* TimingFlow
* TimingBuffered
* TimingSimple
* GaugeFlow
* GaugeBuffered
* GaugeSimple

#### Simple

"Simple" just calculates only min, max, avg and count. It's works quite simple and stupid,
doesn't require extra CPU and/or RAM.

#### Flow

"Flow" calculates min, max, avg, count, per1, per10, per50, per90 and per99 ("per" is a shorthand for "percentile").
It doesn't store observed values (only summarized/aggregated ones)

A percentile value calculation is implemented this way:

### Buffered

???

Func metrics
============

There're also metrics with the "Func" ending:
* GaugeFloat64Func.
* GaugeInt64Func.

This metrics accepts a function as an argument so they call the function
to update their value by themselves.

An example:

```go
server := echo.New()
    
[...]
    
engineInstance := fasthttp.WithConfig(engine.Config{
     Address:      srv.Address,
})

metrics.GaugeInt64Func(
    "concurrent_incoming_connections",
    nil,
    func() int64 { return int64(engineInstance.GetOpenConnectionsCount()) },
).SetGCEnabled(false)

server.Run(engineInstance)
```

Performance
===========

```
BenchmarkList-3                                      	      30	  15311873 ns/op	  989904 B/op	      24 allocs/op
BenchmarkRegistry-3                                  	 2000000	       189 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryReal-3                              	 1000000	       672 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal_lazy-3                     	  500000	      1012 ns/op	     352 B/op	       3 allocs/op
BenchmarkRegistryRealReal_normal-3                   	  500000	       879 ns/op	      16 B/op	       1 allocs/op
BenchmarkRegistryRealReal_FastTags_withHiddenTag-3   	  500000	       934 ns/op	       0 B/op	       0 allocs/op
BenchmarkRegistryRealReal_FastTags-3                 	  500000	       891 ns/op	       0 B/op	       0 allocs/op
```

Developer notes
===============

The structure of a metric object
--------------------------------

To deduplicate code it's used an approach similar to C++'s inheritance, but using Golang's composition. Here's the scheme:
![composition/inheritance](https://raw.githubusercontent.com/trafficstars/metrics/master/docs/implementation_composition.png)

* `registryItem` makes possible to register the metric into the registry
* `common` handles the common (for all metric types) routines like GC or `Sender`.
* `commonAggregative` handles the common routines for all aggregative metrics (like statistics slicing)

Iterators
---------

