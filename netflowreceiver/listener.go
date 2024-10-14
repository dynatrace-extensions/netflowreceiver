package netflowreceiver

import (
	"errors"
	"fmt"
	"net"

	"github.com/netsampler/goflow2/v2/decoders/netflow"
	protoproducer "github.com/netsampler/goflow2/v2/producer/proto"
	"github.com/netsampler/goflow2/v2/utils"
	"github.com/netsampler/goflow2/v2/utils/debug"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type Listener struct {
	config      Config
	logger      *zap.Logger
	recv        *utils.UDPReceiver
	logConsumer consumer.Logs
}

type droppedCallback struct {
}

func newDroppedCallback() *droppedCallback {
	return &droppedCallback{}
}
func (r *droppedCallback) Dropped(pkt utils.Message) {
	// TODO - Implement
}

func NewListener(config Config, logger *zap.Logger, logConsumer consumer.Logs) *Listener {
	return &Listener{config: config, logger: logger, logConsumer: logConsumer}
}

// Start starts the listener, it will start listening for NetFlow packets
// on a UDP socket and decode them according to the specific protocol
func (l *Listener) Start() error {

	l.logger.Info("Setting up receivers for listener", zap.Any("config", l.config))

	cfg := &utils.UDPReceiverConfig{
		Sockets:          l.config.Sockets,
		Workers:          l.config.Workers,
		QueueSize:        l.config.QueueSize,
		Blocking:         false,
		ReceiverCallback: newDroppedCallback(),
	}
	recv, err := utils.NewUDPReceiver(cfg)
	if err != nil {
		return err
	}
	l.recv = recv

	decodeFunc, err := l.buildDecodeFunc()
	if err != nil {
		return err
	}

	l.logger.Info("Start listening for NetFlow", zap.Any("config", l.config))
	if err := l.recv.Start(l.config.Hostname, l.config.Port, decodeFunc); err != nil {
		return err
	}

	go l.handleErrors()

	return nil
}

func (l *Listener) buildDecodeFunc() (utils.DecoderFunc, error) {

	// We used a goflow2 proto producer to produce messages using protobuf format
	protoProducer, err := protoproducer.CreateProtoProducer(nil, protoproducer.CreateSamplingSystem)
	if err != nil {
		return nil, err
	}

	// the otel log producer converts those messages into OpenTelemetry logs
	otelLogsProducer := NewOtelLogsProducer(protoProducer, l.logConsumer)

	cfgPipe := &utils.PipeConfig{
		Producer: otelLogsProducer,
	}

	var decodeFunc utils.DecoderFunc
	var p utils.FlowPipe
	if l.config.Scheme == "sflow" {
		p = utils.NewSFlowPipe(cfgPipe)
	} else if l.config.Scheme == "netflow" {
		p = utils.NewNetFlowPipe(cfgPipe)
	} else if l.config.Scheme == "flow" {
		p = utils.NewFlowPipe(cfgPipe)
	} else {
		return nil, fmt.Errorf("scheme does not exist: %s", l.config.Scheme)
	}

	decodeFunc = p.DecodeFlow
	decodeFunc = debug.PanicDecoderWrapper(decodeFunc)

	return decodeFunc, nil
}

func (l *Listener) handleErrors() {
	for {
		select {
		case err := <-l.recv.Errors():
			if errors.Is(err, net.ErrClosed) {
				l.logger.Info("receiver closed")
				continue
			} else if !errors.Is(err, netflow.ErrorTemplateNotFound) && !errors.Is(err, debug.PanicError) {
				l.logger.Error("receiver error", zap.Error(err))
				continue
			} else if errors.Is(err, debug.PanicError) {
				var pErrMsg *debug.PanicErrorMessage
				if errors.As(err, &pErrMsg) {
					l.logger.Error("panic error", zap.String("panic", pErrMsg.Inner))
				}
				l.logger.Error("receiver panic", zap.Error(err))
				continue
			}
		}
	}
}

func (l *Listener) Shutdown() error {
	if l.recv != nil {
		return l.recv.Stop()
	}
	return nil
}
