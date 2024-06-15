package netflowreceiver

import (
	"encoding/json"
	"github.com/netsampler/goflow2/v2/producer"
	protoproducer "github.com/netsampler/goflow2/v2/producer/proto"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

type OtelLogsProducerWrapper struct {
	wrapped     producer.ProducerInterface
	logConsumer consumer.Logs
}

func (o *OtelLogsProducerWrapper) Produce(msg interface{}, args *producer.ProduceArgs) ([]producer.ProducerMessage, error) {
	flowMessageSet, err := o.wrapped.Produce(msg, args)
	if err != nil {
		return flowMessageSet, err
	}

	log := plog.NewLogs()
	resourceLog := log.ResourceLogs().AppendEmpty()
	resourceLog.Resource().Attributes().PutStr("key", "netflow")
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName("netflow")
	scopeLog.Scope().SetVersion("2.0.0")

	for _, msg := range flowMessageSet {
		// we know msg is ProtoProducerMessage
		protoProducerMessage, ok := msg.(*protoproducer.ProtoProducerMessage)
		if !ok {
			continue
		}

		logRecord := scopeLog.LogRecords().AppendEmpty()

		// Time the receiver received the message
		receivedTime := time.Unix(0, int64(protoProducerMessage.TimeReceivedNs))
		logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(receivedTime))

		// Time the flow started
		startTime := time.Unix(0, int64(protoProducerMessage.TimeFlowStartNs))
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(startTime))

		// The bytes of the message in JSON format
		m := make(map[string]interface{})
		jsonFormatted, err := protoProducerMessage.MarshalJSON()
		if err != nil {
			continue
		}
		err = json.Unmarshal(jsonFormatted, &m)
		if err != nil {
			continue
		}
		m["content"] = string(jsonFormatted)
		err = logRecord.Body().SetEmptyMap().FromRaw(m)
		if err != nil {
			continue
		}
	}

	err = o.logConsumer.ConsumeLogs(nil, log)
	if err != nil {
		return flowMessageSet, err
	}

	return flowMessageSet, nil
}

func (o *OtelLogsProducerWrapper) Close() {
	o.wrapped.Close()
}

func (o *OtelLogsProducerWrapper) Commit(flowMessageSet []producer.ProducerMessage) {
	o.wrapped.Commit(flowMessageSet)
}

func NewOtelLogsProducer(wrapped producer.ProducerInterface, logConsumer consumer.Logs) producer.ProducerInterface {
	return &OtelLogsProducerWrapper{
		wrapped:     wrapped,
		logConsumer: logConsumer,
	}
}
