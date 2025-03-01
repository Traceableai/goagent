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

By default, `goagent` includes the [Traceable filter](./filter/traceable) into the server instrumentation (e.g. HTTP server or GRPC server)
based on the [configuration features](https://github.com/Traceableai/agent-config/blob/main/proto/ai/traceable/agent/config/v1/config.proto#L29).
To run Traceable filter we need to:

First compile the binary using the build tag `traceable_filter`, for example:

```bash
go build -tags 'traceable_filter' -o /path-to-app/myapp
```

Then, copy the library into the same folder as the compiled binary:

```bash
curl -sSL https://raw.githubusercontent.com/Traceableai/goagent/main/filter/traceable/copy-library.sh | bash -s -- /path-to-app
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

For running the server with libtraceable running on a docker based environment

```bash
docker build -f _examples/http-server/Dockerfile .  -t http-server:dev  
docker run -it --rm -v /Users/traceable/config.yaml:/go/src/_examples/http-server/config.yaml  -p 8081:8081 http-server:dev
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

## Running unit tests
The repo uses docker to run some unit tests as libtraceable needs a linux based environment and some unit tests involve it.

Steps:
1. Build the base image
    ```bash
    docker build --build-arg UBUNTU_VERSION=20.04 --build-arg GO_VERSION=1.22.8 --build-arg ARCH=arm64  -f ./filter/traceable/cmd/libtraceable-downloader/Dockerfile.ubuntu.test ./filter/traceable/cmd/libtraceable-downloader -t traceable_goagent_test_base:ubuntu_20.04
    ```
2. Build the test image (this will trigger the tests)
    ```bash
    docker build --build-arg TRACEABLE_GOAGENT_DISTRO_VERSION=ubuntu_20.04 --build-arg ARCH=arm64  -f ./_tests/Dockerfile.test . -t test:dev
    ```

