---
date: "2019-02-12:00:00+02:00"
title: "Metrics"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "practical instructions"
---

# Metrics

Metrics work by exposing a `/metrics` endpoint which can then be accessed by prometheus.

To keep the load on the database minimal, metrics are stored and updated in redis.
The `metrics` package provides several functions to create and update metrics.

{{< table_of_contents >}}

## New metrics

First, define a `const` with the metric key in redis. This is done in `pkg/metrics/metrics.go`.

To expose a new metric, you need to register it in the `init` function inside of the `metrics` package like so:

{{< highlight golang >}}
// Register total user count metric
promauto.NewGaugeFunc(prometheus.GaugeOpts{
    Name: "vikunja_team_count", // The key of the metric. Must be unique.
    Help: "The total number of teams on this instance", // A description about the metric itself.
}, func() float64 {
    count, _ := GetCount(TeamCountKey) // TeamCountKey is the const we defined earlier.
    return float64(count)
})
{{< /highlight >}}

Then you'll need to set the metrics initial value on every startup of vikunja.
This is done in `pkg/routes/routes.go` to avoid cyclic imports.
If metrics are enabled, it checks if a redis connection is available and then sets the initial values.
A convenience function is available if the metric is based on a database struct.

Because metrics are stored in redis, you are responsible to increase or decrease these based on criteria you define.
To do this, use `metrics.UpdateCount(value, key)` where `value` is the amount you want to cange it (you can pass
negative values to decrease it) and `key` it the redis key used to define the metric.

## Using it

A Prometheus config with a Grafana template is available at [our git repo](https://git.kolaente.de/vikunja/monitoring).
