<h1 align="center">
  <img src="https://raw.githubusercontent.com/mono424/go-pts/images/logo.png"><br>
  Go (Pneumatic) Tube System
</h1>


[![Run Tests](https://github.com/mono424/go-pts/actions/workflows/run-tests.yml/badge.svg?branch=main)](https://github.com/mono424/go-pts/actions/workflows/run-tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mono424/go-pts)](https://goreportcard.com/report/github.com/mono424/go-pts)
[![codecov](https://codecov.io/gh/mono424/go-pts/branch/main/graph/badge.svg?token=9VA6CYDXAZ)](https://codecov.io/gh/mono424/go-pts)

go-pts is a websocket channel management library written in Go. It offers a rest-style syntax and easily integrates with various websocket and http frameworks.

# Get Started

1. Install go-pts by using the comand below.

```
go get github.com/mono424/go-pts
```

2. Install the driver for your websocket module.

```
go get github.com/mono424/go-pts-gorilla-connector
```

3. Import to your code.

```go
import (
  "github.com/mono424/go-pts"
  "github.com/mono424/go-pts-gorilla-connector"
)
```
