FROM golang:latest AS builder
RUN go install go.opentelemetry.io/collector/cmd/builder@latest


WORKDIR /src
COPY . /src
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 builder --config builder-config.yaml 

FROM scratch

ARG USER_UID=10001
ARG USER_GID=10001
USER ${USER_UID}:${USER_GID}

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/otelcol-dev/otelcol-dev /otelcol-dev
EXPOSE 4317 55680 55679 2055/udp
ENTRYPOINT ["/otelcol-dev"]
CMD ["--config", "/etc/otel/config.yaml"]