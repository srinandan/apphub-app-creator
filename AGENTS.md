# AGENTS.md

Welcome to `apphub-app-creator`! This guide is designed for AI agents and human contributors working on this codebase. It documents repository architecture, core workflows, design conventions, testing patterns, and rules of engagement.

---

## 1. Overview & Purpose

`apphub-app-creator` is a Go-based CLI utility and HTTP microservice that automatically discovers Google Cloud resources and creates/manages [Google Cloud App Hub](https://cloud.google.com/app-hub/docs/overview) Applications, Services, and Workloads.

Discovery is powered by:
- **Cloud Asset Inventory (CAIS)**: Querying GCP assets by resource labels, Resource Manager tags, asset types, or resource name substring patterns.
- **Cloud Logging**: Querying log entry labels to identify active runtime services and deployments.
- **Kubernetes (GKE)**: Discovering services and workloads organized by Kubernetes namespaces or `app.kubernetes.io/name` labels.
- **Multi-Project Aggregation**: Grouping resources across projects under a GCP Folder into dedicated App Hub applications.

---

## 2. Repository Layout

```
.
├── cmd/
│   └── apphub-app-creator/
│       ├── apphub-app-creator.go       # Main CLI entry point; handles version ldflags & executes root command
│       └── apphub-app-creator_test.go  # Smoke tests for main execution
├── internal/
│   ├── client/                         # Core domain logic & Google Cloud API clients
│   │   ├── apphub.go                   # App Hub API integration (Lookup, Create, Delete apps/services/workloads)
│   │   ├── apphub_iface.go             # Interface definition (appHubClient) for mocking App Hub gRPC calls
│   │   ├── apphub-attributes.go        # Parsing and mapping App Hub attributes (criticality, severity, owners)
│   │   ├── cais.go                     # Cloud Asset Inventory search queries and resource parsing
│   │   ├── client.go                   # Discovery orchestrator (GenerateFromAll, GenerateAppsAssetInventory, etc.)
│   │   ├── logging.go                  # Cloud Logging filter query generator and log parser
│   │   ├── trace.go                    # Cloud Trace query utilities
│   │   ├── go.mod                      # Internal module definition
│   │   └── *_test.go                   # Unit tests with mock clients and table-driven cases
│   ├── clilog/                         # Centralized structured logging wrapper
│   │   ├── clilog.go                   # slog.Logger initialization and singleton provider
│   │   └── go.mod                      # Internal module definition
│   └── cmd/                            # Cobra command tree & HTTP server handlers
│       ├── root.go                     # Root cobra command, global flags (--log-level, --disable-check)
│       ├── cmd.go                      # `apps` command group definition and shared flags (--parent, --locations, etc.)
│       ├── generate.go                 # `apps generate` command implementation and flag validation
│       ├── delete.go                   # `apps delete` command implementation
│       ├── server.go                   # `server` HTTP daemon (Gorilla Mux + CORS + graceful shutdown)
│       ├── generateHandler.go          # HTTP POST /generate request parsing, validation, and execution
│       ├── util.go                     # Helper utilities (project ID parsing, resource validation)
│       ├── go.mod                      # Internal module definition
│       └── *_test.go                   # CLI & HTTP handler tests
├── docs/                               # Cobra-generated CLI documentation & generator script (docs.go)
├── samples/                            # Sample payload JSON files for HTTP API and attributes config
│   ├── attributes.json                 # Sample App Hub attribute payload (owners, criticality, environment)
│   ├── sample1.json                    # Auto-detect generation sample
│   ├── sample2.json                    # Label-based generation sample
│   └── sample3.json                    # Tag-based generation sample
├── Dockerfile                          # Multi-stage container build with distroless non-root image
├── .golangci.yml                       # golangci-lint configuration
├── .goreleaser.yml                     # GoReleaser configuration for release builds & signing
├── go.mod                              # Root Go module (v1.24.4) with internal replace directives
├── go.sum                              # Go module checksums
└── CONTRIBUTING.md                     # Contribution guidelines and CLA details
```

---

## 3. Technology Stack & Dependencies

- **Language**: Go `1.24.4`
- **CLI Framework**: [`github.com/spf13/cobra`](https://github.com/spf13/cobra)
- **HTTP Routing & Middleware**: [`github.com/gorilla/mux`](https://github.com/gorilla/mux), [`github.com/rs/cors`](https://github.com/rs/cors)
- **Logging**: Standard Library [`log/slog`](https://pkg.go.dev/log/slog) encapsulated in `internal/clilog`
- **Google Cloud SDKs**:
  - `cloud.google.com/go/apphub/apiv1`
  - `cloud.google.com/go/asset/apiv1`
  - `cloud.google.com/go/logging`
  - `cloud.google.com/go/resourcemanager/apiv3`
  - `cloud.google.com/go/trace/apiv1`

---

## 4. Key Discovery Modes & Selectors

The tool processes resources based on mutually exclusive selectors:

| Selector Flag / JSON Property | Description | Example CLI Usage |
| :--- | :--- | :--- |
| `--auto-detect` / `autoDetect` | Discovers apps using well-known labels & tags (`app`, `application`, `app.kubernetes.io/name`). | `--auto-detect=true` |
| `--label-key` [value] / `label` | Groups assets sharing a specific GCP resource label key (and optional value). | `--label-key="env" --label-value="prod"` |
| `--tag-key` [value] / `tag` | Groups assets sharing a GCP Resource Manager tag key and optional value. | `--tag-key="cost-center" --tag-value="123"` |
| `--log-label-key` [value] / `logLabel` | Discovers active assets from Cloud Logging entry labels. | `--log-label-key="service_name"` |
| `--per-k8s-namespace` / `perK8sNamespace` | Generates one App Hub application per discovered Kubernetes namespace. | `--per-k8s-namespace=true` |
| `--per-k8s-app-label` / `perK8sAppLabel` | Generates one App Hub application per `app.kubernetes.io/name` label. | `--per-k8s-app-label=true` |
| `--project-keys` / `projectKeys` | Groups all assets across specified projects into a single named application. | `--project-keys="p1" --project-keys="p2" --app-name="my-app"` |
| `--contains` / `contains` | Filters resources whose full resource name contains a substring. | `--contains="frontend"` |

### Common Operational Flags:
- `--report-only` (`action.reportOnly` in JSON): Dry-run mode. Discovers assets and logs/returns the planned App Hub application structure without creating resources or mutating GCP.
- `--attributes` (`options.attributes` in JSON): Path to a JSON file defining Criticality, Severity/Environment, and Owner contacts to apply to created applications.

---

## 5. Development Workflow & Commands

### Prerequisites
- Go `1.24.4+` installed locally.
- Google Cloud SDK (`gcloud`) with Application Default Credentials (ADC):
  ```bash
  gcloud auth application-default login
  ```

### Build Binary
```bash
go build -o apphub-app-creator ./cmd/apphub-app-creator/apphub-app-creator.go
```

### Run Tests
```bash
# Run unit tests across all packages
go test -v ./...

# Run unit tests with code coverage profiling
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Run Linter
The repository enforces linting with `golangci-lint`. All checks must pass with zero issues before pushing:
```bash
golangci-lint run --timeout=4m
```

### Run HTTP Server Locally
```bash
./apphub-app-creator server --port=8080
```

### Docker Build & Run
```bash
# Build image locally
docker build -t apphub-app-creator:latest .

# Run CLI via container
docker run -it --rm \
  -v $HOME/.config/gcloud:/root/.config/gcloud \
  apphub-app-creator:latest apps generate --help
```

---

## 6. Code Style & Engineering Guidelines

When making modifications or adding features, adhere to the following conventions:

### 1. Module Structure & Internal Packages
- Root `go.mod` uses `replace` directives for local submodules:
  - `internal/clilog`
  - `internal/client`
  - `internal/cmd`
- Keep dependencies lean. Prefer Go standard library over third-party packages whenever possible.

### 2. Error Handling
- **Never swallow errors**: Avoid `_ = err` or bare returns on failures.
- **Wrap errors with context**: Always use `fmt.Errorf("action description: %w", err)` to preserve the error chain and provide debug context.
- In HTTP handlers, return well-structured JSON error responses (`ErrorResponse{Error: "..."}`) with appropriate HTTP status codes (400 for bad input, 500 for internal errors).

### 3. Structured Logging
- Use `internal/clilog`'s structured logger:
  ```go
  logger := clilog.GetLogger()
  logger.Info("Discovered resource", "uri", uri, "location", location)
  logger.Error("Failed to lookup App Hub workload", "error", err)
  ```
- Respect configured log levels (`info`, `warn`, `error`, `off`).

### 4. Unit Testing & Fakes
- Write table-driven tests (`*_test.go`) alongside the source code.
- **Never make live network/GCP calls in unit tests.**
- Use the `appHubClient` interface in `internal/client/apphub_iface.go` and override function variables (such as `searchAssetsFunc` and `getAppHubClientFunc`) with test fakes.
- Every new exported function, command flag, or HTTP handler must have accompanying test cases covering both happy path and failure/edge conditions.

### 5. Cobra CLI & Validation
- Mutually exclusive flags must be declared and validated using Cobra's built-in validation helpers (`MarkFlagsMutuallyExclusive`, `MarkFlagsRequiredTogether`, `MarkFlagsOneRequired`) in `init()` as well as in the command's `Args` validation func.
- Set `cmd.SilenceUsage = true` in `RunE` to avoid spamming the help text on runtime API errors.

---

## 7. HTTP API Reference

The server exposes the following endpoints:

### `GET /`
- **Purpose**: Health check endpoint.
- **Response**: `200 OK` with body `OK`.

### `POST /generate`
- **Purpose**: Discovers resources and generates App Hub applications.
- **Headers**: `Content-Type: application/json`
- **Request Body Format**:
  ```json
  {
    "selector": {
      "autoDetect": true
    },
    "scope": {
      "parent": "projects/my-gcp-project",
      "locations": ["us-central1", "global"],
      "managementProject": "my-apphub-host-project"
    },
    "action": {
      "reportOnly": true
    },
    "options": {
      "attributes": {
        "criticality": { "type": "MISSION_CRITICAL" },
        "environment": { "type": "PRODUCTION" },
        "developerOwners": [{ "email": "dev@example.com", "displayName": "Dev Team" }]
      }
    }
  }
  ```
- **Response**: `200 OK` with JSON object containing mapped applications, services, and workloads.

---

## 8. Rules for AI Agents

1. **Verification & Zero Lint Failures**: Always run `golangci-lint run --timeout=4m`, `go test -v ./...`, and `go build ./...` after introducing changes. **Never push or commit changes until `golangci-lint` passes cleanly with zero errors/warnings.**
2. **Minimal Diff**: Implement only what was requested. Avoid unnecessary refactoring or adding unused configuration parameters.
3. **No Hardcoded Credentials**: Do not hardcode project IDs, keys, or credentials. Always respect user-supplied flags/payloads and default ADC credentials.
4. **Documentation**: If adding a new command flag or API parameter, update the relevant markdown files in `docs/` and sample JSON files in `samples/`.
5. **Branch & Pull Request Workflow**: **Never commit directly to the `main` branch.** Always create a GitHub issue first, create a dedicated feature or bugfix branch, push the branch to remote, and open a Pull Request for review.
