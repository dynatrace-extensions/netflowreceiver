// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otel_netflow_receiver

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/dynatrace-extensions/netflowreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id:       component.NewIDWithName(metadata.Type, "defaults"),
			expected: createDefaultConfig(),
		},
		{
			id: component.NewIDWithName(metadata.Type, "one_listener"),
			expected: &Config{
				Scheme:    "netflow",
				Port:      2055,
				Sockets:   1,
				Workers:   1,
				QueueSize: 1000,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "zero_queue"),
			expected: &Config{
				Scheme:    "netflow",
				Port:      2055,
				Sockets:   1,
				Workers:   1,
				QueueSize: 1000,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "zero_queue"),
			expected: &Config{
				Scheme:    "netflow",
				Port:      2055,
				Sockets:   1,
				Workers:   1,
				QueueSize: 1000,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "sflow"),
			expected: &Config{
				Scheme:    "sflow",
				Port:      2055,
				Sockets:   1,
				Workers:   1,
				QueueSize: 1000,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "flow"),
			expected: &Config{
				Scheme:    "flow",
				Port:      2055,
				Sockets:   1,
				Workers:   1,
				QueueSize: 1000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestInvalidConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id  component.ID
		err string
	}{
		{
			id:  component.NewIDWithName(metadata.Type, "invalid_schema"),
			err: "scheme must be one of sflow, netflow, or flow",
		},
		{
			id:  component.NewIDWithName(metadata.Type, "invalid_port"),
			err: "port must be greater than 0",
		},
		{
			id:  component.NewIDWithName(metadata.Type, "zero_sockets"),
			err: "sockets must be greater than 0",
		},
		{
			id:  component.NewIDWithName(metadata.Type, "zero_workers"),
			err: "workers must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			err = component.ValidateConfig(cfg)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}
