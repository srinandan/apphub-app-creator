# apphub-app-creator

[![Go Report Card](https://goreportcard.com/badge/github.com/srinandan/apphub-app-creator)](https://goreportcard.com/report/github.com/srinandan/apphub-app-creator)
[![GitHub release](https://img.shields.io/github/v/release/srinandan/apphub-app-creator)](https://github.com/srinandan/apphub-app-creator/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

`apphub-app-creator` is a command-line utility to generate [Google Cloud App Hub](https://cloud.google.com/app-hub/docs/overview) applications from [Cloud Asset Inventory (CAIS)](https://cloud.google.com/asset-inventory/docs/overview) asset searches.

This tool simplifies the process of creating App Hub applications by allowing you to define them based on existing GCP resource labels, tags or resource names.

## Installation

`apphub-app-creator` is a binary and you can download the appropriate one for your platform from [here](https://github.com/srinandan/apphub-app-creator/releases). Run this script to download & install the latest version (on Linux or Darwin)

```sh
curl -L https://raw.githubusercontent.com/srinandan/apphub-app-creator/main/downloadLatest.sh | sh -
```

or

```sh
docker run -ti --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate --help
```

## Usage

The primary command is `generate`, which creates App Hub applications based on a GCP resource label.

### Prerequisites

* Ensure you have authenticated with Google Cloud CLI:

    ```shell
    gcloud auth login
    gcloud auth application-default login
    ```

* The user or service account running the tool must have the following IAM roles:
  * `apphub.admin` on the App Hub management project.
  * `cloudasset.viewer` on the project where your resources are located.
  * `logging.viewer` on the project where logs are written to.

* Please follow the instructions here to setup on [Host Projects](https://cloud.google.com/app-hub/docs/set-up-app-hub-host-project)

* **OR** Please follow the instructions here to enable a [folder](https://cloud.google.com/app-hub/docs/set-up-app-hub-folder) for Application Management.

### Generate Command

The `generate` command requires the following flags:

* `--parent`: (Required) The scope of CAIS Asset Search. Must be of the format projects/{project} or folders/{folder}.
* `--locations`: (Required) GCP location names to filter CAIS Asset Search (e.g. us-central1).
* `--auto-detect`: (Options) Automatically detect applications using well known identifiers through labels and tags.
* `--label-key`: (Optional) The GCP resource label key to filter resources from Cloud Asset Inventory.
* `--label-value`: (Optional) The GCP resource label value to filter resources from Cloud Asset Inventory. Must be used with `label-key`
* `--log-label-key`: (Optional) The GCP Cloud Logging label key to filter resources from Cloud Logging.
* `--log-label-value`: (Optional) The GCP Cloud Logging label value to filter resources from Cloud Logging. Must be used with `log-label-key`
* `--tag-key`: (Optional) The GCP resource tag key to filter resources from Cloud Asset Inventory.
* `--tag-value`: (Optional) The GCP resource tag value to filter resources from Cloud Asset Inventory. Must be used with `tag-key`
* `--contains`: (Optional) GCP Resources whose name contains the string.
* `--per-k8s-namespace`: (Optional) Create one App Hub application per discovered Kubernetes namespace.
* `--per-k8s-app-label`: (Optional) Create one App Hub application per app.kubernetes.io/name label value.
* `--management-project`: (Optional) App Hub Management Project Id. If parent is set to projects/{project}, then management-project defaults to the same.
* `--attributes-file`: (Optional) Path to a JSON file containing App Hub application attributes.
* `--assets-file`: (Optional) Path to a CSV file containing a list of asset types to search in CAIS.
* `--report-only`: (Optional) Generates a report of discovered assets without creating applications or registering services/workloads.

#### Examples

##### Automatically detect applications

To create App Hub applications based on well known labels and tags:

```shell
docker run -it --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --auto-detect=true
```

##### Generate applications based on label key

To create App Hub applications for all resources in `my-gcp-project` that have the label key `appid`, you would run:

```shell
docker run -it --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --label-key="appid"
```

This will:

1. Search for all resources in `my-gcp-project` with the label key `appid`.
2. For each unique value of the `appid` label key, it will create a new App Hub application.
3. The services and workloads for each application will be populated from the resources that share the same label value.

##### Generate applications based on label key and value

To create App Hub applications for all resources in `my-gcp-project` that have the label key `appid` and value `app1`, you would run:

```shell
docker run -it --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --label-key="appid" \
    --label-value="app1"
```

This will:

1. Search for all resources in `my-gcp-project` with the label key `appid` and value `app1`.
2. It will create a new App Hub application the services and workloads for each application will be populated from the resources that share the same label value.

##### Generate applications from multiple locations

To create App Hub applications for all resources in `my-gcp-project` that have the label key `appid` and deployed in multiple locations, you would run:

```shell
docker run -it --rm ghcr.io/srinandan/apphub-app-creator:latest apps generate \
    --project-id="my-gcp-project" \
    --locations="us-central1" \
    --locations="us-east1" \
    --label-key="appid"
```

This will:

1. Search for all resources in `my-gcp-project` with the label key `appid` in the locations `us-central1` and `us-east1.
2. For each unique value of the `appid` label key, it will create a new App Hub application.
3. The services and workloads for each application will be populated from the resources that share the same label value.

### Delete Command

The `delete` command deletes one or more applications in a given set of locations. The `delete` command requires the following flags:

* `--locations`: (Required) GCP location names to delete applications from (e.g. us-central1).
* `--management-project`: (Required) The project where App Hub is managed.

## How do I verify the binary?

All artifacts are signed by [cosign](https://github.com/sigstore/cosign). We recommend verifying any artifact before using them.

You can use the following public key to verify any `apphub-app-creator` binary with:

```sh
cat cosign.pub
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHHFDIsSzmNuYtsR1R0+nElNG3WuY
asYLL8Ko8vw+CvCcGEV7BuI5vJykMBQWlW3XfDoGtPLQ4oxhuCaK21h07w==
-----END PUBLIC KEY-----

cosign verify-blob --key=cosign.pub --signature apphub-app-creator_<platform>_<arch>.zip.sig apphub-app-creator_<platform>_<arch>.zip
```

Where `platform` can be one of `Darwin`, `Linux` or `Windows` and arch (architecture) can be one of `arm64` or `x86_64`

## How do I verify the container?

All images are signed by [cosign](https://github.com/sigstore/cosign). We recommend verifying any container before using them.

```sh
cat cosign.pub
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHHFDIsSzmNuYtsR1R0+nElNG3WuY
asYLL8Ko8vw+CvCcGEV7BuI5vJykMBQWlW3XfDoGtPLQ4oxhuCaK21h07w==
-----END PUBLIC KEY-----

cosign verify --key=cosign.pub ghcr.io/srinandan/apphub-app-creator:latest
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to contribute to this project.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.txt](LICENSE.txt) file for details.

## Support

This is not an officially supported Google product
