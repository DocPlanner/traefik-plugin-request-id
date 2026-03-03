# Traefik Plugin Request ID

A [Traefik](https://traefik.io) middleware plugin that injects an `X-Request-ID` header with a generated UUIDv4 value into HTTP requests and responses, allowing downstream services to correlate and trace individual requests.

Based on [github.com/mdklapwijk/traefik-plugin-request-id](https://github.com/mdklapwijk/traefik-plugin-request-id).

## Overview

X-Request-ID checks incoming requests for the configured header (default `X-Request-ID`). If the header is absent, the plugin generates a new UUIDv4 value and adds it to both the request (so backends can read it) and the response (so clients and proxies can correlate it). If the header is already present, the plugin passes the request through without modification.

The plugin can be disabled at runtime via configuration, and the header name is fully configurable.

## Configuration

### Static (plugin registration)

```yaml
# traefik.yml
experimental:
  plugins:
    requestid:
      moduleName: github.com/docplanner/requestid
      version: v0.1.0
```

### Dynamic (middleware definition)

#### File provider

```yaml
# dynamic.yaml
http:
  middlewares:
    requestid:
      plugin:
        requestid:
          headerName: "X-Request-ID"
          enabled: true

  routers:
    my-router:
      rule: "PathPrefix(`/`)"
      service: my-service
      middlewares:
        - requestid
```

#### Kubernetes CRD

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: requestid
  namespace: traefik-system
spec:
  plugin:
    requestid:
      headerName: "X-Request-ID"
      enabled: true
```

### Configuration Reference

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `headerName` | `string` | No | `X-Request-ID` | Name of the header to inject |
| `enabled` | `bool` | No | `true` | Whether the plugin is active |

### Behavior

- If the configured header is **absent** from the request, a new UUIDv4 is generated and added to both the request and the response.
- If the configured header is **already present**, the request passes through unchanged.
- When `enabled` is `false`, the plugin is a no-op and all requests pass through unmodified.
- UUIDs are generated using Go's [`github.com/google/uuid`](https://pkg.go.dev/github.com/google/uuid) package (version 4, random).

## Local Development

Clone the repository and use the provided `Makefile`:

```bash
make test    # run unit tests with coverage
make lint    # run golangci-lint (requires golangci-lint installed)
```

### Testing with Docker Compose

A ready-to-use local development setup is included in the `docker/` directory. It loads the plugin from the repo root using Traefik's local plugin mode.

```
docker/
├── docker-compose.yml          # Traefik + whoami backend
└── traefik-config/
    ├── traefik.yaml             # Traefik static config (local plugin registration)
    └── dynamic.yaml             # Dynamic config (middleware rules + routing)
```

Start the stack:

```bash
cd docker
docker compose up
```

Test the plugin:

```bash
# Request without X-Request-ID: plugin generates one
curl -sv http://localhost:8888/ 2>&1 | grep -i x-request-id

# Request with existing X-Request-ID: plugin preserves it
curl -sv -H "X-Request-ID: my-custom-id" http://localhost:8888/ 2>&1 | grep -i x-request-id
```

The Traefik dashboard is available at [http://localhost:9090](http://localhost:9090).

Edit `docker/traefik-config/dynamic.yaml` to change middleware settings. Edit `requestid.go` at the repo root to change plugin logic, then restart the stack with `docker compose restart traefik`.

## Publishing to the Traefik Plugin Catalog

To make the plugin available in the [Traefik Plugins Catalog](https://plugins.traefik.io):

1. Ensure the GitHub repository is **public**.
2. Add the `traefik-plugin` **topic** to the repository.
3. Verify `.traefik.yml` exists at the repo root with valid `testData`.
4. Verify `go.mod` exists at the repo root.
5. Create a **git tag** (e.g. `v0.1.0`).
6. The catalog polls GitHub daily and will pick up the plugin automatically.

## Credits

This plugin is a refactored version of [traefik-plugin-request-id](https://github.com/mdklapwijk/traefik-plugin-request-id) by [@mdklapwijk](https://github.com/mdklapwijk), which was itself based on:

- [github.com/pipe01/plugin-requestid](https://github.com/pipe01/plugin-requestid)
- [github.com/gamblingpro/plugin-requestid](https://github.com/gamblingpro/plugin-requestid)

## License

[Apache 2.0](LICENSE)
