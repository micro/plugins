# Plugins [![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![GoDoc](https://godoc.org/github.com/micro/plugins?status.svg)](https://godoc.org/github.com/micro/plugins) [![Unit Tests](https://github.com/micro/plugins/actions/workflows/ci.yml/badge.svg)](https://github.com/micro/plugins/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/micro/go-plugins)](https://goreportcard.com/report/github.com/micro/plugins) [![Vistors](https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgo-micro%2Fplugins&count_bg=%2379C83D&title_bg=%23555555&icon=github.svg&icon_color=%23E7E7E7&title=Vistors&edge_flat=false)](https://hits.seeyoufarm.com)

Go plugins is a place for community maintained plugins.

## Overview

Micro tooling is built on a powerful pluggable architecture. Plugins can be swapped out with zero code changes.
This repository contains plugins for all micro related tools. Read on for further info.

## Getting Started

- [Plugins](#plugins)
	- [Overview](#overview)
	- [Getting Started](#getting-started)
	- [Contents](#contents)
	- [Usage](#usage)
		- [Env](#env)
		- [Flags](#flags)
		- [Options](#options)
	- [Build Pattern](#build-pattern)
	- [Community](#community)

## Contents

Contents of this repository:

| Directory | Description                                                     |
| --------- | ----------------------------------------------------------------|
| Broker    | PubSub messaging; NATS, NSQ, RabbitMQ, Kafka                    |
| Client    | RPC Clients; gRPC, HTTP                                         |
| Codec     | Message Encoding; BSON, Mercury                                 |
| Micro     | Micro Toolkit Plugins                                           |
| Registry  | Service Discovery; Etcd, Gossip, NATS                           |
| Selector  | Load balancing; Label, Cache, Static                            |
| Server    | RPC Servers; gRPC, HTTP                                         |
| Transport | Bidirectional Streaming; NATS, RabbitMQ                         | 
| Wrapper   | Middleware; Circuit Breakers, Rate Limiting, Tracing, Monitoring|

## Usage

Plugins can be added to go-micro in the following ways. By doing so they'll be available to set via command line args or environment variables.

Import the plugins in a `plugins.go` file

```go
package main

import (
	_ "github.com/micro/plugins/v5/broker/rabbitmq"
	_ "github.com/micro/plugins/v5/registry/kubernetes"
	_ "github.com/micro/plugins/v5/transport/nats"
)
```

Create your service and ensure you call `service.Init`

```go
package main

import (
	"go-micro.dev/v5"
)

func main() {
	service := micro.NewService(
		// Set service name
		micro.Name("my.service"),
	)

	// Parse CLI flags
	service.Init()
}
```

Build your service

```
go build -o service ./main.go ./plugins.go
```

### Env

Use environment variables to set the

```
MICRO_BROKER=rabbitmq \
MICRO_REGISTRY=kubernetes \ 
MICRO_TRANSPORT=nats \ 
./service
```

### Flags

Or use command line flags to enable them

```shell
./service --broker=rabbitmq --registry=kubernetes --transport=nats
```

### Options

Import and set as options when creating a new service

```go
import (
	"go-micro.dev/v5"
	"github.com/micro/plugins/v5/registry/kubernetes"
)

func main() {
	registry := kubernetes.NewRegistry() //a default to using env vars for master API

	service := micro.NewService(
		// Set service name
		micro.Name("my.service"),
		// Set service registry
		micro.Registry(registry),
	)
}
```

## Build Pattern

An anti-pattern is modifying the `main.go` file to include plugins. Best practice recommendation is to include
plugins in a separate file and rebuild with it included. This allows for automation of building plugins and
clean separation of concerns.

Create file plugins.go

```go
package main

import (
	_ "github.com/micro/plugins/v5/broker/rabbitmq"
	_ "github.com/micro/plugins/v5/registry/kubernetes"
	_ "github.com/micro/plugins/v5/transport/nats"
)
```

Build with plugins.go

```shell
go build -o service main.go plugins.go
```

Run with plugins

```shell
MICRO_BROKER=rabbitmq \
MICRO_REGISTRY=kubernetes \
MICRO_TRANSPORT=nats \
service
```

## Community

- [Discord](https://discord.gg/qV3HvnEJfB)
- [Slack](https://join.slack.com/t/go-micro/shared_invite/zt-175aaev1d-iHExPTlfxvfkOeeKLIYEYw)
- [QQ Group](https://jq.qq.com/?_wv=1027&k=5Gmrfv9i)
