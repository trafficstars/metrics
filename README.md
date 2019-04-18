[![Build Status](https://travis-ci.org/trafficstars/metrics.svg?branch=master)](https://travis-ci.org/trafficstars/metrics)
[![go report](https://goreportcard.com/badge/github.com/trafficstars/metrics)](https://goreportcard.com/report/github.com/trafficstars/metrics)
[![GoDoc](https://godoc.org/github.com/trafficstars/metrics?status.svg)](https://godoc.org/github.com/trafficstars/metrics)

Description
===========

This is a implementation of high performance handy metrics library for Golang which could be
used for prometheus (passive export) and/or StatsD (active export). But the primary method is
the passive export (a special page where somebody get fetch all the metrics).

How to use
==========

Count the number of requests (and request rate by measuring the rate of the count):
```go
metrics.Count(`requests`, metrics.Tags{
    `method`: request.Method,
}).Increment()
```

Measure the latency:
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

func (sender *statsdSender) SendUint64(metric metrics.Metric, key string, uint64) error {
[... send the metric to statsd ...]
}

func (sender *statsdSender) SendFloat64(metric metrics.Metric, key string, float64) error {
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

(the buffer should be implemented on the sender side if it's required)

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

### Slicing

An aggregative metric has aggregative/summarized statistics for a few periods at the same time:
* `Last` -- is the very last value ever received via `ConsiderValue`.
* `Current` -- is the statistics for the current second (which is not complete, yet)
* `1S` -- is the statistics for the previous second
* `5S` -- is the statistics for the previous 5 seconds
* `1M` -- is the statistics for the previous minute 
* ...
* `6H` -- is the statistics for the last 6 hours
* `1D` -- is the statistics for the last day
* `Total` -- is the total statistics 

Once per second the `Current` became `1S` and an empty `Current` appears. And there's a history of the last
5 statistics for `1S` which is used to recalculate statistics for `5S`. There's in turn a history of the last
12 statistics for `5S` which is used to recalculate statistics for `1M`. And so on.

This process is called "slicing" (which is done once per second by default).

To change aggregation periods and slicing interval you can use methods `SetAggregationPeriods` and `SetSlicerInterval`
accordingly.

A note: So if you have one aggregative metric it will export every value (max, count, ...) for every aggregation period
(`Total`, `Last`, `Current`, `1S`, `5S`, ...).

### Aggregation types

If you have no time to read how every aggregation type works then just read "Use case"
of every type.

#### Simple

"Simple" just calculates only min, max, avg and count. It's works quite simple and stupid,
doesn't require extra CPU and/or RAM.

###### Use case

Any case where it's not required to get percentile values.

#### Flow

"Flow" calculates min, max, avg, count, per1, per10, per50, per90 and per99 ("per" is a shorthand for "percentile").
It doesn't store observed values (only summarized/aggregated ones)

###### Use case

* It's required to get percentile values, but they could be inaccurate.
* There's a lot of values per second.
* There will be a lot of such metrics.

###### How the calculation of percentile values works

It just increases/decreases the value (let it call "P") to reach required ratio of [values lower than the value "P"] to [values
higher than the value "P"].

Let's image `ConsiderValue` was called. We do not store previous values so we:

1. Pick a random number `[0..1)`. If it's less than the required percentile then we think that this value should be
   lower than the current value (and vice versa).
2. Correct the current value if the prediction in the first stage was wrong.

The function (that implements the above algorithm) is called `guessPercentile` (see `common_aggregative_flow.go`).

There's a constant `iterationsRequiredPerSecond` to tune accuracy of the algorithm. The more this constant value is the
more accurate is the algorithm, but more values is required (to be passed through `ConsiderValue`) per second to
approach the real value. It's set to `20`, so this kind of aggregative metrics shouldn't be used if the next condition
is not satisfied: `VPS >> 20` (`VPS` means "values per second", `>>` means "much more than").

![flow](https://raw.githubusercontent.com/trafficstars/metrics/master/internal/docs/demonstration/flow/flow.png)

The more values are passed the more inert is the value and the more accurate it is.
So, again, the "Flow" method should be used only on high `VPS`.

*Attention!* There's an unsolved problem of correct merging percentile-related statistics:
For example, to calculate percentile statistics for interval "5 seconds" it's required to merge statistics
for 5 different seconds (with their-own percentile values), so the resulting
percentile value is calculated as just the weighted average of percentile values. It's correct only if the load
is monotone. Otherwise it will be inaccurate, but *usually* good enough.

#### Buffered

"Buffered" calculates min, max, avg, count and stores values samples to be able to
calculate any percentile values at any time. This method more precise than the "Flow", but requires much more RAM. The
size of the buffer with the sample values is regulated via method `SetAggregativeBufferSize`
(the default value is "1000"); the more buffer size is the more accuracy of percentile values is,
but more RAM and CPU is required.

###### Use case

* It's required to get precise percentile values
* There won't be a lot of such metrics (otherwise it will utilize a lot of RAM).

###### Buffer handling

There're two buffer-related specifics:
* The buffer is limited, so how do we handle the rest events (if there're more events per second
than the buffer size).
* How are two buffers get merged to the new one of the same size (see "slicing").

Both problems are solved using the same initial idea:
Let's imagine we received a 1001-th value (via `ConsiderValue`), while our buffer
is only 1000 elements long. Then we just:
* Skip it with probability 1/1001.
* If it's not skipped then it override a random element of the buffer.

If we receive a second element, then we skip it with probability 2/1002... And so on.

It's proven that it's any event value will have an equal probability to get into the buffer.
And 1000 elements is enough to calculate value of percentile 99 (there will be 10 element with a higher value). 

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
![composition/inheritance](https://raw.githubusercontent.com/trafficstars/metrics/master/internal/docs/scheme/implementation_composition.png)

* `registryItem` makes possible to register the metric into the registry
* `common` handles the common (for all metric types) routines like GC or `Sender`.
* `commonAggregative` handles the common routines for all aggregative metrics (like statistics slicing)

Iterators
---------

