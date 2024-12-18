// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otel_netflow_receiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/dynatrace-extensions/netflowreceiver/internal/metadata"
)

const (
	defaultSockets   = 1
	defaultWorkers   = 2
	defaultQueueSize = 1_000
)

// NewFactory creates a factory for netflow receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, metadata.LogsStability))
}

// Config defines configuration for netflow receiver.
// By default we listen for netflow traffic on port 2055
func createDefaultConfig() component.Config {
	return &Config{
		Scheme:    "netflow",
		Port:      2055,
		Sockets:   defaultSockets,
		Workers:   defaultWorkers,
		QueueSize: defaultQueueSize,
	}
}

// createLogsReceiver creates a netflow receiver.
// We also create the UDP receiver, which is the piece of software that actually listens
// for incoming netflow traffic on an UDP port.
func createLogsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Logs) (receiver.Logs, error) {
	conf := *(cfg.(*Config))

	nr, err := newNetflowLogsReceiver(params, conf, consumer)
	if err != nil {
		return nil, err
	}

	return nr, nil
}
