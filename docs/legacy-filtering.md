# Legacy Filtering
This page describes the old style of filtering and is deprecated and removed in
agent version 5.0+. See [Filtering](filtering.md) for how to configure
filtering in SignalFx Smart Agent 4.7.0+.

## Old-style inclusion list filtering
In the Smart Agent prior to version 4.7.0, custom metrics were filtered out of
the agent by means of a `whitelist.json` file that was referenced under the
`metricsToExclude` section of the standard distributed config file.  This
method is now deprecated but will continue to work if already configured as
such.  If you upgrade to 4.7.0 or later, you will continue to need the
`whitelist.json` custom-metric filtering until you set `enableBuiltInFiltering:
true` in your config.  You should **remove** the reference to `whitelist.json` in
your config at the same time you set that flag.

Essentially the effect of setting `enableBuiltInFiltering: true` and removing
the reference to `whitelist.json` is to make the monitors inherently emit the
metrics that were in `whitelist.json` by default, which makes it significantly
easier to configure monitors to emit additional metrics in a single step.

We recommend upgrading the agent to at least 4.7.0 and setting the
`enableBuiltInFitlering: true` flag, as it is much easier to configure and
understand than the old `metricsToExlude`/`metricsToInclude` mechanism.  The
only reason to not set `enableBuiltInFiltering: true` is if you have extensive
modifications to the allowed metrics (especially via `metricsToInclude`)
and you don't want to rewrite all of that using `extraMetrics` (the new
built-in filtering will filter out metrics before they can be processed by
`metricsToInclude`, so that config will have no effect if built-in filtering is
enabled).

## Global datapoint filtering

**DEPRECATED** -- The following information is about filtering
datapoints at a global level (i.e. outside of monitor configuration).  It will
continue to work, but we recommend putting filter definitions [at the monitor
level](#additional-monitor-level-filtering) instead.

Filters can be specified with the `metricsToExclude` config at the top-level of
the agent config.  This parameter should be a list of filters that specify the
metric name and/or dimensions to match on.  These filters should be used
sparingly, as every single datapoint emitted by monitors must be processed by
these filters, resulting in higher CPU usage.

Metric names and dimension values can be either globbed (i.e. where `*` is a
wildcard for zero or more characters, and `?` is a wildcard for a single
character) or specified as a Go-compatible regular expression (the value must
be surrounded by `/` to be considered a regex).

You can limit the scope of a filter to a specific monitor type with the
`monitorType` option in the filter item.  This value matches the monitor type
used in the monitor config.  This is very useful when trying to filter on a
heavily overloaded dimension, such as the `plugin_instance` dimension that most
collectd-based monitors emit.

Sometimes it is easier to filter the metrics you want to allow through.
You can do this by setting the `negated` option to `true` on a filter item.
This makes all metric name and dimension matching negated so that only
datapoints with that name or dimension are allowed through.  You would also
use this option to specify custom metrics that you want to send to SignalFx.
The `monitorType` option is never negated, and always serves to scope a filter
item to the specified monitor type, if given. In the case where multiple filters
overlap, they will be combined by the agent. If metrics are matched by two filters
which have opposite negation, the agent will favor exclusion.

Examples:

```yaml
  # If a metric matches any one of these filters it will be excluded.
  metricsToExclude:
    # Excludes all metrics that start with 'container_spec_memory_'
    - metricName: container_spec_memory_*

    # Excludes multiple metric names in a single filter
    - metricNames:
      - container_start_time_seconds
      - container_tasks_state

    # Excludes networking metrics about flannel interfaces that match the
    # pattern (e.g. 'flannel.1', 'flannel.2', etc.)
    - metricName: pod_network_*
      dimensions:
        interface: /flannel\.\d/

    # This is a metric from collectd.  All instances of it will be suppressed.
    - metricName: vmpage_io.swap.in

    # This filter will only match against datapoints from the collectd/df
    # monitor
    - dimensions:
        plugin_instance: dm-*
      monitorType: collectd/df

    # This only allows cpu and memory metrics from the docker monitor to come
    # through.  Note that the monitorType is never negated, and metrics from
    # other monitors will not be excluded since this filter is scoped to a
    # single monitor.
    - metricNames:
      - cpu*
      - memory*
      monitorType: docker-container-stats
      negated: true

    # This indicates that you want to monitor the custom metrics
    # gauge.cluster.status and gauge.thread_pool.active for the
    # elasticsearch monitor.
    - metricNames:
      - elasticsearch.cluster.status
      - elasticsearch.thread_pool.active
      monitorType: elasticsearch
      negated: true

    # This will be automatically merged with the above filter to produce one
    # Filter on three metric names for elasticsearch
    - metricNames:
      - elasticsearch.thread_pool.inactive
      monitorType: elasticsearch
      negated: true

    # This will override the above filter, exclusion is always favored
    - metricNames:
      - elasticsearch.thread_pool.inactive
      monitorType: elasticsearch
```

### Inclusion filtering
Sometimes it is necessary to specify metrics that you want to include even when
they would be otherwise excluded by filters in `metricsToExclude`.  Therefore,
there is the `metricsToInclude` option that should be a list of the same type
of metric filters used in `metricsToExclude`.

The main difference between a filter in `metricsToExclude` that has `negated:
true` and a filter in `metricsToInclude` is that the filter in
`metricsToExclude` will actually exclude metrics that don't match the filter,
whereas the filter in `metricsToInclude` will never exclude anything, but only
override the exclusion in `metricsToExclude`.

For example:

```yaml
  metricsToExclude:
   - metricNames:
     - request_size
     - response_size
  metricsToInclude:
   # This allows through datapoints with the name `request_size` that have the
   # specified dimension but still excludes other datapoints without those
   # dimensions.  All `response_size` metrics are still excluded.
   - metricName: request_size
     dimensions:
       app: bigapp
```

This can be useful for overridding the built-in inclusion list for metrics.
