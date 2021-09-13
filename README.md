# Go Agent

`goagent` provides a set of instrumentation features for collecting relevant tracing data as well as secure an application by blocking requests selectively.

## Getting started

Setting up Go Agent can be done with a few lines:

```go
import "github.com/Traceableai/goagent/config"

...

func main() {
    cfg := config.Load()
    cfg.ServiceName = config.String("myservice")

    shutdown := goagent.Init(cfg)
    defer shutdown
}
```

Config values can be declared in config file, env variables or code. For further information about config check [this section](config/README.md).

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
        fooHandler,
        "/foo/{bar}",
    ))

    // ...
}
```

#### Options

##### Filter

Filtering can be added as part of options. Multiple filters can be added and they will be run in sequence until a filter returns true (request is blocked), or all filters are run.

```go

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
go run ./examples/http-client/main.go
```

In terminal 2 run the server:

```bash
go run ./examples/http-server/main.go
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
go run ./examples/gin-server/main.go
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

Filtering can be added as part of options. Multiple filters can be added and they will be run in sequence until a filter returns true (request is blocked), or all filters are run.

```go

// ...

    grpc.UnaryInterceptor(
        traceablegrpc.UnaryServerInterceptor(
            traceablegrpc.WithFilter(filter.NewMultiFilter(filter1, filter2))
        ),
    ),

// ...

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
go run ./examples/grpc-client/main.go
```

In terminal 2 run the server:

```bash
go run ./examples/grpc-server/main.go
```

## Other instrumentations

- [database/traceablesql](instrumentation/database/traceablesql)
- [github.com/gorilla/traceablemux](instrumentation/github.com/gorilla/traceablemux)

## Contributing

### Running tests

Tests can be run with (requires docker)

```bash
make test
```

for unit tests only

```bash
make test-unit
```

### Releasing

Run `./release.sh <version_number>` (`<version_number>` should follow semver, e.g. `1.2.3`). The script will change the hardcoded version, commit it, push a tag and prepare the hardcoded version for the next release.
