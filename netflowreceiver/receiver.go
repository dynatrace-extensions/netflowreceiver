package netflowreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type netflowReceiver struct {
	host        component.Host
	cancel      context.CancelFunc
	config      *Config
	logConsumer consumer.Logs
	logger      *zap.Logger
	listeners   []*Listener
}

func (nr *netflowReceiver) Start(ctx context.Context, host component.Host) error {
	nr.host = host
	ctx = context.Background()
	ctx, nr.cancel = context.WithCancel(ctx)

	// The receiver configuration is composed of a list of listeners
	// Each listener process a specific flow protocol (NetFlow v5, NetFlow v9, IPFIX, sFlow)
	// and listens on a specific address and UDP port
	for _, listenerConfig := range nr.config.Listeners {
		listener := NewListener(listenerConfig, nr.logger, nr.logConsumer)
		if err := listener.Start(); err != nil {
			return err
		}
		nr.listeners = append(nr.listeners, listener)
	}

	nr.logger.Info("NetFlow receiver started")
	return nil
}

func (nr *netflowReceiver) Shutdown(ctx context.Context) error {
	nr.logger.Info("NetFlow receiver is shutting down")
	for _, listener := range nr.listeners {
		err := listener.Shutdown()
		if err != nil {
			nr.logger.Error("Error shutting down listener", zap.Error(err))
		}
	}
	nr.cancel()
	return nil
}
