# Akira MQTT

[![build](https://github.com/gsalomao/akira/actions/workflows/build.yml/badge.svg)](https://github.com/gsalomao/akira/actions/workflows/build.yml/badge.svg)
[![codecov](https://codecov.io/gh/gsalomao/akira/branch/master/graph/badge.svg?token=VlcHlx3NAc)](https://codecov.io/gh/gsalomao/akira)
[![Go Report Card](https://goreportcard.com/badge/github.com/gsalomao/akira)](https://goreportcard.com/report/github.com/gsalomao/akira)
[![license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> **_NOTE:_**  This project is under development, DO NOT use it in production.

Akira is an open-source, embeddable, and high-performance MQTT server, compliant with the MQTT 3.1, 3.1.1, and 5.0
specification.

This project is an [Apache 2.0](./LICENSE) licensed MQTT server developed in [Go](https://go.dev/).

#### What is MQTT?

MQTT stands for MQ Telemetry Transport. It is a publish-subscribe, extremely simple and lightweight messaging protocol,
designed for constrained devices and low-bandwidth, high-latency or unreliable networks.
[Learn more](https://mqtt.org/faq)

### Features

- [ ] Fully compatible with MQTT 3.1, 3.1.1 and 5.0 specifications
	- Packet Properties
    - Topic Aliases
    - Shared Subscriptions
    - Subscription Options and Subscription Identifiers
    - Message Expiry
    - Client Session Expiry
    - QoS Control Quotas
    - Server-side Disconnect and Auth Packets
    - Will Delay Intervals
    - $SYS topics
    - Retained messages
- [ ] MQTT over TCP, TLS, WebSocket and Secure WebSocket
- [ ] Embeddable
- [ ] Extensible through hooks

## Contributing

Please follow the
[Contributing Guide](https://github.com/gsalomao/akira/blob/master/CONTRIBUTING.md)

## License

This project is released under
[Apache 2.0 License](https://github.com/gsalomao/akira/blob/master/LICENSE).
