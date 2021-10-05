# Go Agent

`goagent` provides a set of instrumentation features for collecting relevant tracing data as well as secure an application by blocking requests selectively using Traceable features.

## Getting started

Setting up Go Agent can be done with a few lines:

```go
import (
    "github.com/Traceableai/goagent"
    "github.com/Traceableai/goagent/config"
)

//...

func main() {
    cfg := config.Load()
    cfg.Tracing.ServiceName = config.String("myservice")

    shutdown := goagent.Init(cfg)
    defer shutdown()
}
```

Config values can be declared in config file, env variables or code. For further information about config check [this section](config/README.md).

### Traceable filter

By default, `goagent` includes the [Traceable filter](./filter/traceable) into server instrumentation (e.g. HTTP server or GRPC server)
based on the [configuration features](https://github.com/Traceableai/agent-config/blob/main/proto/ai/traceable/agent/config/v1/config.proto#L29).
To run Traceable filter we need to:

Fist compile the binary using the build tag `traceable_filter`, for example:

```bash
go build -tags 'traceable_filter' -o myapp
```

Then, we need to download the library next to the application binary:

```bash
# Install libtraceable downloader (run this from a non go.mod folder)
go install github.com/Traceableai/goagent/filter/traceable/cmd/libtraceable-downloader@latest

...

# Pull library in the binary directory
cd /path/to/myapp &&
libtraceable-downloader pull-library
```

## Package net/traceablehttp

### HTTP server

The server instrumentation relies on the `http.Handler` component of the server declarations.

```go
import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"
)

func main() {
    // ...

    r := mux.NewRouter()
    r.Handle("/foo/{bar}", traceablehttp.NewHandler(
        fooHandler, // existing user handler
        "/foo/{bar}", // name of the span generated for this handler
    ))

    // ...
}
```

#### Options

##### Filter

Filters allow users to filter requests based on URL, headers or body. Filters can be added in the handler declaration using the `traceablehttp.WithFilter` option. Multiple filters can be added too by using `filter.NewMultiFilter` and they will be run in sequence until a filter returns true (request is blocked), or all filters are run.

```go
import "github.com/hypertrace/goagent/sdk/filter"

// ...

    r.Handle("/foo/{bar}", traceablehttp.NewHandler(
        fooHandler,
        "/foo/{bar}",
        traceablehttp.WithFilter(filter.NewMultiFilter(filter1, filter2)),
    ))

// ...

````

### HTTP client

The client instrumentation relies on the `http.Transport` component of the HTTP client in Go.

```go
import (
    "net/http"
    "github.com/Traceableai/goagent/instrumentation/net/traceablehttp"
)

// ...

client := http.Client{
    Transport: traceablehttp.NewTransport(
        http.DefaultTransport,
    ),
}

req, _ := http.NewRequest("GET", "http://example.com", nil)

res, err := client.Do(req)

// ...
```

### Running HTTP examples

In terminal 1 run the client:

```bash
go run ./_examples/http-client/main.go
```

In terminal 2 run the server:

```bash
go run ./_examples/http-server/main.go
```

## Gin-Gonic Server

Gin server instrumentation relies on adding the `traceablegin.Middleware` middleware to the gin server.

```go
r := gin.Default()

cfg := config.Load()
cfg.ServiceName = config.String("http-gin-server")

flusher := goagent.Init(cfg)
defer flusher()

r.Use(traceablegin.Middleware())
```

To run an example gin server with the middleware:

```bash
go run ./_examples/gin-server/main.go
```

Then make a request to `localhost:8080/ping`

## Package google.golang.org/traceablegrpc

### GRPC server

The server instrumentation relies on the `grpc.UnaryServerInterceptor` component of the server declarations.

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(
        traceablegrpc.UnaryServerInterceptor(),
    ),
)
```

#### Options

##### Filter

Filters allow users to filter requests based on URL, headers or body. Filters can be added in the server interceptor declaration using the `traceablegrpc.WithFilter` option. Multiple filters can be added too by using `filter.NewMultiFilter` and they will be run in sequence until a filter returns true (request is blocked), or all filters are run.

```go
import "github.com/hypertrace/goagent/sdk/filter"

// ...

    grpc.UnaryInterceptor(
        traceablegrpc.UnaryServerInterceptor(
            traceablegrpc.WithFilter(filter.NewMultiFilter(filter1, filter2))
        ),
    ),
````

### GRPC client

The client instrumentation relies on the `http.Transport` component of the HTTP client in Go.

```go
import (
    // ...

    traceablegrpc "github.com/Traceableai/goagent/instrumentation/google.golang.org/traceablegrpc"
    "google.golang.org/grpc"
)

func main() {
    // ...
    conn, err := grpc.Dial(
        address,
        grpc.WithInsecure(),
        grpc.WithBlock(),
        grpc.WithUnaryInterceptor(
            traceablegrpc.UnaryClientInterceptor(),
        ),
    )
    if err != nil {
        log.Fatalf("could not dial: %v", err)
    }
    defer conn.Close()

    client := pb.NewCustomClient(conn)

    // ...
}
```

### Running GRPC examples

In terminal 1 run the client:

```bash
go run ./_examples/grpc-client/main.go
```

In terminal 2 run the server:

```bash
go run ./_examples/grpc-server/main.go
```

## Other instrumentations

- [database/traceablesql](instrumentation/database/traceablesql)
- [github.com/gorilla/traceablemux](instrumentation/github.com/gorilla/traceablemux)
