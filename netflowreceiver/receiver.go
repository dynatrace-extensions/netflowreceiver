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
	config      Config
	logConsumer consumer.Logs
	logger      *zap.Logger
	listener    *Listener
}

func (nr *netflowReceiver) Start(ctx context.Context, host component.Host) error {
	nr.host = host
	ctx = context.Background()
	ctx, nr.cancel = context.WithCancel(ctx)

	listener := NewListener(nr.config, nr.logger, nr.logConsumer)
	if err := listener.Start(); err != nil {
		return err
	}
	nr.listener = listener
	nr.logger.Info("NetFlow receiver started")
	return nil
}

func (nr *netflowReceiver) Shutdown(ctx context.Context) error {
	nr.logger.Info("NetFlow receiver is shutting down")
	err := nr.listener.Shutdown()
	if err != nil {
		return err
	}
	nr.cancel()
	return nil
}
