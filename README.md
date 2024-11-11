# Netflow Receiver

The netflow receiver is capable of listening for [netflow](https://en.wikipedia.org/wiki/NetFlow), [sflow](https://en.wikipedia.org/wiki/SFlow) or [IPFIX](https://en.wikipedia.org/wiki/IP_Flow_Information_Export) UDP traffic and generating log entries based on the flow content.

This gives Opentelemetry users the capability of monitoring network traffic, and answer questions like:

* Which protocols are passing through the network?
* Which servers and clients are producing the highest amount of traffic?
* What ports are involved in these network calls?
* How many bytes and packets are being sent and received?

The receiver listens for flows and decodes them using the templates that are sent by the flow producers. The data then is converted to JSON and produces structured log records.

## Using the receiver

[Build a collector](https://opentelemetry.io/docs/collector/custom-collector/) with the receiver by using:

```yaml
receivers:
  - gomod: github.com/dynatrace-extensions/netflowreceiver v1.2.0
```

Example `builder-config.yaml`:

```yaml
dist:
  name: otelcol-dev
  description: Basic OTel Collector distribution for Developers
  output_path: ./otelcol-dev

exporters:
  - gomod:
      go.opentelemetry.io/collector/exporter/debugexporter v0.99.0
  - gomod:
      go.opentelemetry.io/collector/exporter/otlphttpexporter v0.99.0

processors:
  - gomod:
      go.opentelemetry.io/collector/processor/batchprocessor v0.99.0

receivers:
  - gomod: github.com/dynatrace-extensions/netflowreceiver v1.2.0
    path: .
```

## Project structure

The receiver code is under `netflowreceiver`.

The folder `otelcol-dev` contains a generated collector built via the `builder-config.yaml` file using [ocb](https://opentelemetry.io/docs/collector/custom-collector/).

There is an example `config.yaml` file used to run the collector with `--config config.yaml`


## Scaling

https://opentelemetry.io/docs/collector/scaling/

