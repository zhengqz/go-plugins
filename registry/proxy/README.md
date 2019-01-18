# Proxy Registry

This is a registry plugin for the micro [proxy](https://micro.mu/docs/proxy.html)

## Usage

Here's a simple usage guide

### Run Proxy

```
go get github.com/micro/micro
```

```
micro proxy
```

### Import and Flag plugin

```
import _ "github.com/micro/go-plugins/registry/proxy"
```

```
go run main.go --registry=proxy
```
