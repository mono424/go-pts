<h1 align="center">
  <img src="https://raw.githubusercontent.com/mono424/go-pts/images/logo.png"><br>
  Go (Pneumatic) Tube System
</h1>


[![Run Tests](https://github.com/mono424/go-pts/actions/workflows/run-tests.yml/badge.svg?branch=main)](https://github.com/mono424/go-pts/actions/workflows/run-tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mono424/go-pts)](https://goreportcard.com/report/github.com/mono424/go-pts)
[![codecov](https://codecov.io/gh/mono424/go-pts/branch/main/graph/badge.svg?token=9VA6CYDXAZ)](https://codecov.io/gh/mono424/go-pts)

Go-PTS is a flexible package for managing Pub-Sub over WebSockets in Go. It offers a rest-style syntax and easily integrates with various websocket and http frameworks.

## Installation

### Installing the main library

1. Get the `go-pts` package using the following command:

```shell
go get github.com/mono424/go-pts
```

### Using connectors

To use go-pts with a specific websocket library, you need to install the corresponding connector.

```shell
go get github.com/mono424/go-pts-gorilla-connector
```

Then, you can import them in your code like this:

```go
import (
  "github.com/mono424/go-pts"
  ptsc_gorilla "github.com/mono424/go-pts-gorilla-connector"
)
```

## Client Libraries

For client-side integration, you can use one of the following client libraries:

| Language | URL |
| -------- | --- |
| JavaScript | [go-pts-client-js](https://github.com/mono424/go-pts-client-js) |
| Dart | [go-pts-client-dart](https://github.com/mono424/go-pts-client-dart) |

## Connectors

For server-side integration with WebSocket libraries, you can use one of the following connectors:

| WebSocket Library | URL |
| ----------------- | --- |
| Gorilla WebSocket | [go-pts-gorilla-connector](https://github.com/mono424/go-pts-gorilla-connector) |
| Melody | [go-pts-melody-connector](https://github.com/mono424/go-pts-melody-connector) |

## Getting Started

1. Create a new TubeSystem

```go
tubeSystem := pts.New(ptsc_gorilla.NewConnector(
  websocket.Upgrader{},
  func(err *pts.Error) {
    println(err.Description)
  },
))
```

2. Register Channels

```go
tubeSystem.RegisterChannel("/stream/:streamId", pts.ChannelHandlers{
  OnSubscribe: func(s *pts.Context) {
    println("Client joined: " + s.FullPath)
  },
  OnMessage: func(s *pts.Context, message *pts.Message) {
    println("New Message on " + s.FullPath + ": " + string(message.Payload))
  },
  OnUnsubscribe: func(s *pts.Context) {
    println("Client left: " + s.FullPath)
  },
})
```

3. Provide a connect route

```go
r.GET("/connect", func(c *gin.Context) {
  properties := make(map[string]interface{}, 1)
  properties["ctx"] = c

  if err := tubeSystem.HandleRequest(c.Writer, c.Request, properties); err != nil {
    println("Something went wrong while handling a Socket request")
  }
})
```

4. Connect from a frontend lib
```javascript
const client = new GoPTSClient({ url: socketUrl, debugging: true })
client.subscribeChannel("test", console.log);
client.send("test", { payload: { foo: "bar" } })
```

## Examples

To get a quick overview of how to use Go-PTS, check out the `examples` folder.
