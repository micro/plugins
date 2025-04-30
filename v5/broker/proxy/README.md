# Proxy Broker

This is a broker plugin for the micro [proxy](https://micro.mu/docs/proxy.html)

## Usage

Here's a simple usage guide

### Run Proxy

```
# install mu
go install github.com/micro/micro/v5/mu@latest

# run proxy
mu proxy
```

### Import and Flag plugin

```
import _ "github.com/micro/plugins/v5/broker/proxy"
```

```
go run main.go --broker=proxy
```
