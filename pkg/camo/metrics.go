// Copyright (c) 2012-2019 Eli Janssen
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package camo provides an HTTP proxy server with content type
// restrictions as well as regex host allow list support.
package camo

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Namespace used for Prometheus metrics.
const MetricNamespace = "camo"
const MetricSubsystem = "proxy"

var (
	contentLengthExceeded = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      "content_length_exceeded_total",
			Help:      "The number of requests where the content length was exceeded.",
		},
	)
	responseFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      "reponses_failed_total",
			Help:      "The number of responses that failed to send to the client.",
		},
	)
	responseTruncated = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Subsystem: MetricSubsystem,
			Name:      "reponses_truncated_total",
			Help:      "The number of responess that were too large to send.",
		},
	)
)
