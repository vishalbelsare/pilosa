// Copyright 2022 Molecula Corp. (DBA FeatureBase).
// SPDX-License-Identifier: Apache-2.0
package statsd

import (
	"sort"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/featurebasedb/featurebase/v3/logger"
	"github.com/featurebasedb/featurebase/v3/stats"
)

// StatsD protocol wrapper using the DataDog library that added Tags to the StatsD protocol
// statsD defailt host is "127.0.0.1:8125"

const (
	// bufferLen Stats lient buffer size.
	bufferLen = 1024
)

// Ensure client implements interface.
var _ stats.StatsClient = &statsClient{}

// statsClient represents a StatsD implementation of pilosa.statsClient.
type statsClient struct {
	client *statsd.Client
	tags   []string
	logger logger.Logger

	// prefix is appended to each metric event name
	prefix string
}

// NewStatsClient returns a new instance of StatsClient.
func NewStatsClient(host string, namespace string) (*statsClient, error) {
	c, err := statsd.NewBuffered(host, bufferLen)
	if err != nil {
		return nil, err
	}

	return &statsClient{
		client: c,
		logger: logger.NopLogger,
		prefix: namespace + ".",
	}, nil
}

// Open no-op
func (c *statsClient) Open() {}

// Close closes the connection to the agent.
func (c *statsClient) Close() error {
	return c.client.Close()
}

// Tags returns a sorted list of tags on the client.
func (c *statsClient) Tags() []string {
	return c.tags
}

// WithTags returns a new client with additional tags appended.
func (c *statsClient) WithTags(tags ...string) stats.StatsClient {
	return &statsClient{
		client: c.client,
		tags:   unionStringSlice(c.tags, tags),
		logger: c.logger,
	}
}

// Count tracks the number of times something occurs per second.
func (c *statsClient) Count(name string, value int64, rate float64) {
	if err := c.client.Count(c.prefix+name, value, c.tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Count error: %s", err)
	}
}

// CountWithCustomTags tracks the number of times something occurs per second with custom tags.
func (c *statsClient) CountWithCustomTags(name string, value int64, rate float64, t []string) {
	tags := append(c.tags, t...)
	if err := c.client.Count(c.prefix+name, value, tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Count error: %s", err)
	}
}

// Gauge sets the value of a metric.
func (c *statsClient) Gauge(name string, value float64, rate float64) {
	if err := c.client.Gauge(c.prefix+name, value, c.tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Gauge error: %s", err)
	}
}

// Histogram tracks statistical distribution of a metric.
func (c *statsClient) Histogram(name string, value float64, rate float64) {
	if err := c.client.Histogram(c.prefix+name, value, c.tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Histogram error: %s", err)
	}
}

// Set tracks number of unique elements.
func (c *statsClient) Set(name string, value string, rate float64) {
	if err := c.client.Set(c.prefix+name, value, c.tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Set error: %s", err)
	}
}

// Timing tracks timing information for a metric.
func (c *statsClient) Timing(name string, value time.Duration, rate float64) {
	if err := c.client.Timing(c.prefix+name, value, c.tags, rate); err != nil {
		c.logger.Errorf("statsd.StatsClient.Timing error: %s", err)
	}
}

// SetLogger sets the logger for client.
func (c *statsClient) SetLogger(logger logger.Logger) {
	c.logger = logger
}

// unionStringSlice returns a sorted set of tags which combine a & b.
func unionStringSlice(a, b []string) []string {
	// Sort both sets first.
	sort.Strings(a)
	sort.Strings(b)

	// Find size of largest slice.
	n := len(a)
	if len(b) > n {
		n = len(b)
	}

	// Exit if both sets are empty.
	if n == 0 {
		return nil
	}

	// Iterate over both in order and merge.
	other := make([]string, 0, n)
	for len(a) > 0 || len(b) > 0 {
		if len(a) == 0 {
			other, b = append(other, b[0]), b[1:]
		} else if len(b) == 0 {
			other, a = append(other, a[0]), a[1:]
		} else if a[0] < b[0] {
			other, a = append(other, a[0]), a[1:]
		} else if b[0] < a[0] {
			other, b = append(other, b[0]), b[1:]
		} else {
			other, a, b = append(other, a[0]), a[1:], b[1:]
		}
	}
	return other
}
