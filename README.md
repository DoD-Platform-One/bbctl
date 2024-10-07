<!-- Warning: Do not manually edit this file. See notes on gluon + helm-docs at the end of this file for more information. -->
# bbctl

![Version: 0.7.4-bb.0](https://img.shields.io/badge/Version-0.7.4--bb.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.7.4](https://img.shields.io/badge/AppVersion-0.7.4-informational?style=flat-square)

bbctl as a helm chart for partial automated management of Big Bang.

## Introduction

`bbctl` is a command line interface (CLI) tool to simplify development, deployment, auditing, and maintaining the deployment of Big Bang a kubernetes cluster.

## User Guide

Follow the [user guide](/docs/user-guide.md) for how to install and use the `bbctl` tool.

## Developer Documentation

Help Contribute! See the [developer documentation](/docs/developer.md). The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

## `bbctl` Usage and Design Priorities

### Automated usage over interactive usage

`bbctl` is primarily intended to be piped to/from other tools and shell scripts. Interactive use is a future possibility.

### Multiple execution contexts

`bbctl` will be running as a cronjob in cluster, possibly as web server in cluster, potentially in pipelines, and on developer machines.

### External _and_ internal users

`bbctl` is currently used both inside and outside the Big Bang team as a fully open source project.

## Upstream References
* <https://repo1.dso.mil/big-bang/product/packages/bbctl>

* <https://repo1.dso.mil/big-bang/product/packages/bbctl>

### Upstream Release Notes

There is no upstream for this chart.

## Learn More
* [Application Overview](docs/overview.md)
* [Other Documentation](docs/)

## Pre-Requisites

* Kubernetes Cluster deployed
* Kubernetes config installed in `~/.kube/config`
* Helm installed

Install Helm

https://helm.sh/docs/intro/install/

## Deployment

* Clone down the repository
* cd into directory
```bash
helm install bbctl chart/
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| bigbang | object | `{"addons":{"authservice":{"enabled":false,"values":{"selector":{"key":"protect","value":"keycloak"}}}},"domain":"bigbang.dev","istio":{"enabled":false,"hardened":{"enabled":false}},"monitoring":{"enabled":false},"networkPolicies":{"controlPlaneCidr":"0.0.0.0/0","controlPlaneNode":null,"enabled":false},"openshift":false}` | Passdown values from Big Bang |
| image.repository | string | `"registry1.dso.mil/ironbank/big-bang/bbctl"` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.tag | string | `""` |  |
| yqImage.repository | string | `"registry1.dso.mil/ironbank/opensource/yq/yq"` |  |
| yqImage.pullPolicy | string | `"Always"` |  |
| yqImage.tag | string | `"4.44.3"` |  |
| imagePullSecrets[0].name | string | `"private-registry"` |  |
| nameOverride | string | `""` |  |
| fullnameOverride | string | `""` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.name | string | `""` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| securityContext.runAsGroup | int | `1000` |  |
| resources | object | `{}` |  |
| nodeSelector | object | `{}` |  |
| tolerations | list | `[]` |  |
| affinity | object | `{}` |  |
| istio.enabled | bool | `false` |  |
| baseConfig.bbctl-log-add-source | bool | `true` |  |
| baseConfig.bbctl-log-format | string | `"json"` |  |
| baseConfig.bbctl-log-level | string | `"warn"` |  |
| baseConfig.bbctl-log-output | string | `"stderr"` |  |
| baseConfig.big-bang-credential-helper | string | `"credentials-file"` |  |
| baseConfig.big-bang-repo | string | `"https://repo1.dso.mil/big-bang/bigbang/-/blob/master/"` |  |
| baseConfig.output-config.format | string | `"yaml"` |  |
| baseLabels | object | `{}` |  |
| bigbangUpdater.enabled | bool | `true` |  |
| bigbangUpdater.schedule | string | `"0 * * * *"` |  |
| bigbangUpdater.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangUpdater.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangUpdater.labels | object | `{}` |  |
| bigbangUpdater.config | object | `{}` |  |
| bigbangUpdater.podAnnotations | object | `{}` |  |
| bigbangUpdater.serviceAccount.create | bool | `true` |  |
| bigbangUpdater.serviceAccount.annotations | object | `{}` |  |
| bigbangUpdater.serviceAccount.name | string | `""` |  |

## Contributing

Please see the [contributing guide](./CONTRIBUTING.md) if you are interested in contributing.

---

_This file is programatically generated using `helm-docs` and some BigBang-specific templates. The `gluon` repository has [instructions for regenerating package READMEs](https://repo1.dso.mil/big-bang/product/packages/gluon/-/blob/master/docs/bb-package-readme.md)._

