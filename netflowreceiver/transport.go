package netflowreceiver

import (
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

type ConsoleTransport struct {
}

func (c *ConsoleTransport) Prepare() error {
	return nil
}

func (c *ConsoleTransport) Init() error {
	return nil
}

func (c *ConsoleTransport) Close() error {
	return nil
}

func (c *ConsoleTransport) Send(key, data []byte) error {
	fmt.Println(string(data))
	return nil
}

type LogConsumerTransport struct {
	logConsumer consumer.Logs
}

func (l *LogConsumerTransport) Prepare() error {
	return nil
}

func (l *LogConsumerTransport) Init() error {
	return nil
}

func (l *LogConsumerTransport) Close() error {
	return nil
}

func (l *LogConsumerTransport) Send(key []byte, data []byte) error {
	log := plog.NewLogs()
	resourceLog := log.ResourceLogs().AppendEmpty()
	resourceLog.Resource().Attributes().PutStr("key", string(key))
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()

	scopeLog.Scope().SetName("netflow")  // TODO - Check what this needs to be
	scopeLog.Scope().SetVersion("1.0.0") // TODO - Check what this needs to be

	logRecord := scopeLog.LogRecords().AppendEmpty()

	// TODO - Extract from the sample
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	logRecord.SetSeverityNumber(plog.SeverityNumberInfo)
	logRecord.SetSeverityText(plog.SeverityNumberInfo.String())

	// TODO - Extract other attributes like timestamp, severity, etc.
	// TODO - unmarshal this to an actual struct ?
	m := make(map[string]interface{})
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	m["content"] = string(data)
	err = logRecord.Body().SetEmptyMap().FromRaw(m)
	if err != nil {
		return err
	}
	err = l.logConsumer.ConsumeLogs(nil, log)
	if err != nil {
		return err
	}

	return nil
}
