# apphub-app-creator

[![CI](https://github.com/srinandan/apphub-app-creator/actions/workflows/ci.yml/badge.svg)](https://github.com/srinandan/apphub-app-creator/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/srinandan/apphub-app-creator)](https://github.com/srinandan/apphub-app-creator/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/srinandan/apphub-app-creator.svg)](https://pkg.go.dev/github.com/srinandan/apphub-app-creator)
[![Go Version](https://img.shields.io/github/go-mod/go-version/srinandan/apphub-app-creator)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`apphub-app-creator` is a CLI utility and HTTP service to automatically discover Google Cloud resources and manage [Google Cloud App Hub](https://cloud.google.com/app-hub/docs/overview) Applications, Services, and Workloads using [Cloud Asset Inventory (CAIS)](https://cloud.google.com/asset-inventory/docs/overview), Cloud Logging, Kubernetes metadata, or resource naming conventions.

---

## Installation

### Binary Download (Linux / macOS)
Download and install the latest binary using the automated install script:

```sh
curl -L https://raw.githubusercontent.com/srinandan/apphub-app-creator/main/downloadLatest.sh | sh -
```

Alternatively, download platform-specific binaries directly from the [Releases](https://github.com/srinandan/apphub-app-creator/releases) page.

### Docker Container
Run directly via the pre-built container image:

```sh
docker run -ti --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate --help
```

---

## Prerequisites

* Ensure you have authenticated with Google Cloud CLI:
  ```shell
  gcloud auth login
  gcloud auth application-default login
  ```

* The user or service account executing the tool must possess the following IAM roles:
  * `roles/apphub.admin` on the App Hub management project.
  * `roles/cloudasset.viewer` on the project or folder where resources reside.
  * `roles/logging.viewer` on the project where Cloud Logging entries are stored.

* Ensure App Hub is set up on a [Host Project](https://cloud.google.com/app-hub/docs/set-up-app-hub-host-project) or enabled for a [Folder](https://cloud.google.com/app-hub/docs/set-up-app-hub-folder).

---

## Discovery Modes & Selectors

The tool discovers and groups resources into App Hub applications based on mutually exclusive selectors:

| Selector Flag | HTTP API Field | Description | Example |
| :--- | :--- | :--- | :--- |
| `--auto-detect` | `selector.autoDetect` | Discovers apps using well-known labels & tags (`app`, `application`, `app.kubernetes.io/name`). | `--auto-detect=true` |
| `--label-key` `[--label-value]` | `selector.label` | Groups assets sharing a specific GCP resource label key (and optional value). | `--label-key="env" --label-value="prod"` |
| `--tag-key` `[--tag-value]` | `selector.tag` | Groups assets sharing a GCP Resource Manager tag key and optional value. | `--tag-key="cost-center" --tag-value="123"` |
| `--log-label-key` `[--log-label-value]` | `selector.logLabel` | Discovers active assets from Cloud Logging entry labels. | `--log-label-key="service_name"` |
| `--per-k8s-namespace` | `selector.perK8sNamespace` | Generates one App Hub application per discovered Kubernetes namespace. | `--per-k8s-namespace=true` |
| `--per-k8s-app-label` | `selector.perK8sAppLabel` | Generates one App Hub application per `app.kubernetes.io/name` label. | `--per-k8s-app-label=true` |
| `--project-keys` `--app-name` | `selector.projectKeys` | Groups all assets across specified projects into a single named application. | `--project-keys="p1" --app-name="my-app"` |
| `--contains` | `selector.contains` | Filters resources whose full resource name contains a substring. | `--contains="frontend"` |

---

## CLI Usage

For complete documentation on all flags and subcommands, see the [CLI Documentation](./docs/apphub-app-creator.md).

### `apps generate` Examples

#### 1. Auto-Detect Applications
Automatically discover applications based on standard application labels and tags:

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --auto-detect=true
```

#### 2. Group by Resource Label Key
Create one App Hub application for each distinct value of a given label key (e.g. `appid`):

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --label-key="appid"
```

#### 3. Match Specific Label Key and Value
Group resources matching both a specific label key and value into a single application:

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --label-key="appid" \
    --label-value="app1"
```

#### 4. Discovered Across Multiple Locations
Discover and associate resources across multiple regions into App Hub:

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --locations="us-east1" \
    --label-key="appid"
```

#### 5. Dry-Run / Preview (`--report-only`)
Simulate discovery and print planned App Hub applications, services, and workloads without modifying GCP resources:

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --auto-detect=true \
    --report-only=true
```

#### 6. Apply Metadata Attributes (`--attributes`)
Attach Criticality, Environment/Severity, and Contact Owner metadata defined in a JSON file:

```shell
./apphub-app-creator apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --auto-detect=true \
    --attributes="samples/attributes.json"
```

### `apps delete`

Delete one or more applications across specified locations:

```shell
./apphub-app-creator apps delete \
    --management-project="my-apphub-host-project" \
    --locations="us-central1" \
    --locations="global" \
    --app-name="my-app"
```

---

## Server Mode & HTTP API

`apphub-app-creator` can run as an HTTP microservice daemon for programmatic integrations.

### Starting the Server

```shell
./apphub-app-creator server --port=8080
```

### Endpoints

#### `GET /healthz`
Liveness endpoint. Returns `200 OK` with body `OK`.

#### `GET /`
In the container image (built with the `embedui` tag) this serves the embedded web UI. In the plain CLI binary, where no UI is embedded, `/` is a health endpoint that returns `200 OK` with body `OK`.

#### `POST /generate`
Discovers resources and generates App Hub applications based on a JSON request payload.

**Example Request:**
```shell
curl -X POST http://localhost:8080/generate \
    -H "Content-Type: application/json" \
    -d @samples/sample1.json
```

**Example Payload (`sample1.json`):**
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

---

## Web UI (Frontend)

The repository includes a modern Vue 3 web interface located in [`frontend/`](./frontend/) to interactively configure selectors, preview discovery results, and manage App Hub applications.

### Running the Web UI Locally

1. Start the backend server:
   ```shell
   ./apphub-app-creator server --port=8080
   ```

2. Start the frontend development server:
   ```shell
   cd frontend
   npm install
   npm run dev
   ```

3. Open your browser at `http://localhost:5173`.

### Running the Bundled UI (Container Image)

The published container image embeds the production (minified) build of the UI and serves it directly from the server — no separate frontend deployment is required:

```shell
docker run -ti --rm -p 8080:8080 ghcr.io/srinandan/apphub-app-creator:latest server --port=8080
```

Then open `http://localhost:8080`. The UI calls the API from the same origin (`/generate`), and `/healthz` provides a liveness check. Live discovery still requires Google Cloud credentials (Application Default Credentials) available to the container.

---

## AI Agent Skill

This repository provides an [Agent Skill](https://agentskills.io) packaged in [`apphub-app-creator-skill/`](./apphub-app-creator-skill/) that teaches AI coding assistants (GitHub Copilot CLI, Gemini CLI, Claude Code, etc.) how to discover Google Cloud resources and manage App Hub applications, services, and workloads.

### Installing via GitHub CLI (`gh skill`)

Install the skill directly into your environment using the GitHub CLI:

```shell
gh skill install srinandan/apphub-app-creator apphub-app-creator-skill
```

To preview the skill before installing:

```shell
gh skill preview srinandan/apphub-app-creator apphub-app-creator-skill
```

### Using with Local Agents

Once installed (or by copying [`apphub-app-creator-skill/SKILL.md`](./apphub-app-creator-skill/SKILL.md) to your workspace or agent skills directory such as `~/.agents/skills/apphub-app-creator-skill/` or `~/.gemini/config/skills/apphub-app-creator-skill/`), your AI assistant can:
- **Analyze infrastructure** and select the optimal discovery mode (`--auto-detect`, `--label-key`, `--tag-key`, `--per-k8s-namespace`, etc.).
- **Perform safe audits** by running `--report-only` first before creating cloud resources.
- **Attach governance metadata** (business criticality, environment classification, owner contacts) using formatted `attributes.json` configs.
- **Manage lifecycle** of App Hub applications, services, and workloads automatically.

---

## Artifact Verification

All release binaries and container images are cryptographically signed using [Cosign](https://github.com/sigstore/cosign).

### Verify Binary

```sh
cat cosign.pub
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHHFDIsSzmNuYtsR1R0+nElNG3WuY
asYLL8Ko8vw+CvCcGEV7BuI5vJykMBQWlW3XfDoGtPLQ4oxhuCaK21h07w==
-----END PUBLIC KEY-----

cosign verify-blob --key=cosign.pub --signature apphub-app-creator_<platform>_<arch>.zip.sig apphub-app-creator_<platform>_<arch>.zip
```

Where `<platform>` is `Darwin`, `Linux`, or `Windows` and `<arch>` is `arm64` or `x86_64`.

### Verify Container

```sh
cosign verify --key=cosign.pub ghcr.io/srinandan/apphub-app-creator:latest
```

---

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) and [AGENTS.md](AGENTS.md) for contribution guidelines, architecture details, and development conventions.

---

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.txt](LICENSE.txt) file for details.

---

## Support

This is not an officially supported Google product.
