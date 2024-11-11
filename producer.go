package otel_netflow_receiver

import (
	"context"
	"encoding/json"
	"github.com/netsampler/goflow2/v2/producer"
	protoproducer "github.com/netsampler/goflow2/v2/producer/proto"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"net/netip"
	"time"
)

var (
	etypeName = map[uint32]string{
		0x806:  "ARP",
		0x800:  "IPv4",
		0x86dd: "IPv6",
	}
	protoName = map[uint32]string{
		1:   "ICMP",
		6:   "TCP",
		17:  "UDP",
		58:  "ICMPv6",
		132: "SCTP",
	}

	flowTypeName = map[int32]string{
		0: "UNKNOWN",
		1: "SFLOW_5",
		2: "NETFLOW_V5",
		3: "NETFLOW_V9",
		4: "IPFIX",
	}
)

type NetworkAddress struct {
	Address string `json:"address,omitempty"`
	Port    uint32 `json:"port,omitempty"`
}

type Flow struct {
	Type           string `json:"type,omitempty"`
	TimeReceived   uint64 `json:"time_received,omitempty"`
	Start          uint64 `json:"start,omitempty"`
	End            uint64 `json:"end,omitempty"`
	SequenceNum    uint32 `json:"sequence_num,omitempty"`
	SamplingRate   uint64 `json:"sampling_rate,omitempty"`
	SamplerAddress string `json:"sampler_address,omitempty"`
}

type Protocol struct {
	Name []byte `json:"name,omitempty"` // Layer 7
}

type NetworkIO struct {
	Bytes   uint64 `json:"bytes,omitempty"`
	Packets uint64 `json:"packets,omitempty"`
}

type OtelNetworkMessage struct {
	Source      NetworkAddress `json:"source,omitempty"`
	Destination NetworkAddress `json:"destination,omitempty"`
	Transport   string         `json:"transport,omitempty"` // Layer 4
	Type        string         `json:"type,omitempty"`      // Layer 3
	IO          NetworkIO      `json:"io,omitempty"`
	Flow        Flow           `json:"flow,omitempty"`
}

func getEtypeName(etype uint32) string {
	if name, ok := etypeName[etype]; ok {
		return name
	}
	return "unknown"
}

func getProtoName(proto uint32) string {
	if name, ok := protoName[proto]; ok {
		return name
	}
	return "unknown"
}

func getFlowTypeName(flowType int32) string {
	if name, ok := flowTypeName[flowType]; ok {
		return name
	}
	return "unknown"
}

// OtelLogsProducerWrapper is a wrapper around a producer.ProducerInterface that sends the messages to a log consumer
type OtelLogsProducerWrapper struct {
	wrapped     producer.ProducerInterface
	logConsumer consumer.Logs
}

// Produce converts the message into a list log records and sends them to log consumer
func (o *OtelLogsProducerWrapper) Produce(msg interface{}, args *producer.ProduceArgs) ([]producer.ProducerMessage, error) {
	flowMessageSet, err := o.wrapped.Produce(msg, args)
	if err != nil {
		return flowMessageSet, err
	}

	log := plog.NewLogs()
	resourceLog := log.ResourceLogs().AppendEmpty()
	resourceLog.Resource().Attributes().PutStr("key", "netflow")
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName("netflow-receiver")
	scopeLog.Scope().SetVersion("1.0.0")

	for _, msg := range flowMessageSet {
		// we know msg is ProtoProducerMessage
		pm, ok := msg.(*protoproducer.ProtoProducerMessage)
		if !ok {
			continue
		}

		srcAddr, _ := netip.AddrFromSlice(pm.SrcAddr)
		dstAddr, _ := netip.AddrFromSlice(pm.DstAddr)
		samplerAddr, _ := netip.AddrFromSlice(pm.SamplerAddress)

		otelMessage := OtelNetworkMessage{
			Source: NetworkAddress{
				Address: srcAddr.String(),
				Port:    pm.SrcPort,
			},
			Destination: NetworkAddress{
				Address: dstAddr.String(),
				Port:    pm.DstPort,
			},
			Transport: getProtoName(pm.Proto), // Layer 4
			Type:      getEtypeName(pm.Etype), // Layer 3
			IO: NetworkIO{
				Bytes:   pm.Bytes,
				Packets: pm.Packets,
			},
			Flow: Flow{
				Type:           getFlowTypeName(int32(pm.Type)),
				TimeReceived:   pm.TimeReceivedNs,
				Start:          pm.TimeFlowStartNs,
				End:            pm.TimeFlowEndNs,
				SequenceNum:    pm.SequenceNum,
				SamplingRate:   pm.SamplingRate,
				SamplerAddress: samplerAddr.String(),
			},
		}

		logRecord := scopeLog.LogRecords().AppendEmpty()

		// Time the receiver received the message
		receivedTime := time.Unix(0, int64(pm.TimeReceivedNs))
		logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(receivedTime))

		// Time the flow started
		startTime := time.Unix(0, int64(pm.TimeFlowStartNs))
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(startTime))

		// The bytes of the message in JSON format
		m, err := json.Marshal(otelMessage)
		if err != nil {
			continue
		}

		// Convert to a map[string]
		// https://opentelemetry.io/docs/specs/otel/logs/data-model/#type-mapstring-any
		sec := map[string]interface{}{}
		if err = json.Unmarshal(m, &sec); err != nil {
			continue
		}

		err = logRecord.Body().SetEmptyMap().FromRaw(sec)
		if err != nil {
			continue
		}
	}

	err = o.logConsumer.ConsumeLogs(context.TODO(), log)
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
