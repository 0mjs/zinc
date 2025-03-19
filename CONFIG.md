# Zinc Configuration Options

Zinc provides a flexible configuration system that allows you to customize the behavior of your web application. You can configure various aspects such as server settings, routing behavior, timeouts, and more.

## Basic Usage

```go
app := zinc.New(zinc.Config{
    DefaultAddr: "127.0.0.1:3000",
    ServerHeader: "MyApp",
    AppName: "My Zinc Application",
})
```

## Complete Configuration Reference

### Server Configuration

| Property     | Type   | Description                                            | Default        |
| ------------ | ------ | ------------------------------------------------------ | -------------- |
| DefaultAddr  | string | The default address the server will listen on          | "0.0.0.0:6530" |
| ServerHeader | string | Sets the value of the Server HTTP header               | "Zinc"         |
| AppName      | string | The name of the application (used in startup messages) | ""             |

### Timeout Settings

| Property        | Type          | Description                                                          | Default |
| --------------- | ------------- | -------------------------------------------------------------------- | ------- |
| ShutdownTimeout | time.Duration | Maximum duration to wait for server shutdown                         | 10s     |
| ReadTimeout     | time.Duration | Maximum duration for reading the entire request                      | 5s      |
| WriteTimeout    | time.Duration | Maximum duration before timing out writes of the response            | 10s     |
| IdleTimeout     | time.Duration | Maximum time to wait for the next request when keep-alive is enabled | 120s    |

### Router Settings

| Property       | Type  | Description                                                       | Default |
| -------------- | ----- | ----------------------------------------------------------------- | ------- |
| CaseSensitive  | bool  | When enabled, /Foo and /foo are different routes                  | false   |
| StrictRouting  | bool  | When enabled, /api and /api/ are different routes                 | false   |
| BodyLimit      | int64 | Maximum allowed size for a request body in bytes                  | 4MB     |
| Concurrency    | int   | Maximum number of concurrent connections                          | 256K    |
| RouteCacheSize | int   | Number of routes stored in the route cache (0 to disable caching) | 1000    |

### Proxy Settings

| Property                | Type     | Description                                                       | Default           |
| ----------------------- | -------- | ----------------------------------------------------------------- | ----------------- |
| EnableTrustedProxyCheck | bool     | Enables checking if a proxy is trusted when determining client IP | false             |
| TrustedProxies          | []string | List of trusted proxy IP addresses or CIDR blocks                 | []                |
| ProxyHeader             | string   | Header to use for client IP addresses when behind a trusted proxy | "X-Forwarded-For" |

### Miscellaneous

| Property                  | Type | Description                                           | Default |
| ------------------------- | ---- | ----------------------------------------------------- | ------- |
| DisableKeepalive          | bool | Disables keep-alive connections                       | false   |
| DisableDefaultContentType | bool | Disables the default Content-Type header in responses | false   |
| DisableStartupMessage     | bool | Disables the startup message when the server starts   | false   |
| EnablePrintRoutes         | bool | Enables printing all registered routes on startup     | false   |

## Examples

### Configuring Timeouts

```go
app := zinc.New(zinc.Config{
    ReadTimeout: 30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout: 60 * time.Second,
    ShutdownTimeout: 15 * time.Second,
})
```

### Configuring Routing Behavior

```go
app := zinc.New(zinc.Config{
    CaseSensitive: true,  // /Users and /users are different routes
    StrictRouting: true,  // /api and /api/ are different routes
})
```

### Configuring Server Headers

```go
app := zinc.New(zinc.Config{
    ServerHeader: "MyAwesomeAPI/1.0",
    DisableDefaultContentType: true,  // Don't set Content-Type header automatically
})
```

### Configuring Proxy Handling

```go
app := zinc.New(zinc.Config{
    EnableTrustedProxyCheck: true,
    TrustedProxies: []string{
        "127.0.0.1",
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
    },
    ProxyHeader: "X-Forwarded-For",
})
```

### Configuring Performance

```go
app := zinc.New(zinc.Config{
    BodyLimit: 50 * 1024 * 1024,  // 50MB max request size
    Concurrency: 10000,           // Handle up to 10K concurrent connections
    RouteCacheSize: 5000,         // Cache up to 5K routes
})
```

## Setting Configuration Programmatically

You can also set the configuration after creating the app:

```go
app := zinc.New()
config := zinc.Config{
    DefaultAddr: "127.0.0.1:3000",
    ServerHeader: "CustomServer",
}
app.SetConfig(&config)
``` 