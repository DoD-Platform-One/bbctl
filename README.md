<!-- Warning: Do not manually edit this file. See notes on gluon + helm-docs at the end of this file for more information. -->
# bbctl

![Version: 2.0.0-bb.1](https://img.shields.io/badge/Version-2.0.0--bb.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.2.0](https://img.shields.io/badge/AppVersion-1.2.0-informational?style=flat-square) ![Maintenance Track: bb_integrated](https://img.shields.io/badge/Maintenance_Track-bb_integrated-green?style=flat-square)

bbctl as a helm chart for partial automated management of Big Bang.

## Introduction

`bbctl` is a command line interface (CLI) tool to simplify development, deployment, auditing, and maintaining the deployment of Big Bang a kubernetes cluster.

This repository is for the `bbctl` chart. For `bbctl` application, development please refer to: https://repo1.dso.mil/big-bang/apps/developer-tools/bbctl.

## Developer Documentation

Help Contribute! See the [developer documentation](/docs/developer.md). The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

## `bbctl` Usage and Design Priorities

### Multiple execution contexts

`bbctl` will be running as a cronjob in cluster, possibly as web server in cluster, potentially in pipelines, and on developer machines.

### External _and_ internal users

`bbctl` is currently used both inside and outside the Big Bang team as a fully open source project.

## Upstream References

- <https://repo1.dso.mil/big-bang/product/packages/bbctl>
- <https://repo1.dso.mil/big-bang/product/packages/bbctl>

## Upstream Release Notes

There is no upstream for this chart.
Application development for `bbctl` is done in this repository: https://repo1.dso.mil/big-bang/apps/developer-tools/bbctl.

## Learn More

- [Application Overview](docs/overview.md)
- [Other Documentation](docs/)

## Pre-Requisites

- Kubernetes Cluster deployed
- Kubernetes config installed in `~/.kube/config`
- Helm installed

Install Helm

https://helm.sh/docs/intro/install/

## Deployment

- Clone down the repository
- cd into directory

```bash
helm install bbctl chart/
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| domain | string | `"dev.bigbang.mil"` | Passdown values from Big Bang These values are for the BigBang overrides |
| networkPolicies.enabled | bool | `false` |  |
| networkPolicies.controlPlaneCidr | string | `"0.0.0.0/0"` |  |
| networkPolicies.ingressLabels.app | string | `"istio-ingressgateway"` |  |
| networkPolicies.ingressLabels.istio | string | `"ingressgateway"` |  |
| networkPolicies.additionalPolicies | list | `[]` |  |
| istio.enabled | bool | `false` |  |
| istio.hardened.enabled | bool | `false` |  |
| istio.hardened.outboundTrafficPolicyMode | string | `"REGISTRY_ONLY"` |  |
| istio.hardened.customServiceEntries[0].name | string | `"external-service-entries-bbctl"` |  |
| istio.hardened.customServiceEntries[0].enabled | bool | `true` |  |
| istio.hardened.customServiceEntries[0].spec.hosts[0] | string | `"repo1.dso.mil"` |  |
| istio.hardened.customServiceEntries[0].spec.location | string | `"MESH_EXTERNAL"` |  |
| istio.hardened.customServiceEntries[0].spec.ports[0].number | int | `443` |  |
| istio.hardened.customServiceEntries[0].spec.ports[0].protocol | string | `"TLS"` |  |
| istio.hardened.customServiceEntries[0].spec.ports[0].name | string | `"https"` |  |
| istio.hardened.customServiceEntries[0].spec.resolution | string | `"DNS"` |  |
| istio.hardened.customAuthorizationPolicies | list | `[]` |  |
| istio.hardened.clusterWideHardenedEnabled | bool | `false` |  |
| monitoring.enabled | bool | `false` |  |
| bbtests.enabled | bool | `false` |  |
| image.repository | string | `"registry1.dso.mil/ironbank/big-bang/bbctl"` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.tag | string | `"1.2.0"` |  |
| yqImage.repository | string | `"registry1.dso.mil/ironbank/opensource/yq/yq"` |  |
| yqImage.pullPolicy | string | `"Always"` |  |
| yqImage.tag | string | `"4.45.4"` |  |
| registryCredentials.registry | string | `"registry1.dso.mil"` |  |
| registryCredentials.username | string | `""` |  |
| registryCredentials.password | string | `""` |  |
| registryCredentials.email | string | `""` |  |
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
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"128Mi"` |  |
| resources.limits.cpu | string | `"200m"` |  |
| resources.limits.memory | string | `"256Mi"` |  |
| nodeSelector | object | `{}` |  |
| tolerations | list | `[]` |  |
| affinity | object | `{}` |  |
| credentialsFile.credentials[0].uri | string | `"registry1.dso.mil"` |  |
| credentialsFile.credentials[0].username | string | `""` |  |
| credentialsFile.credentials[0].password | string | `""` |  |
| credentialsFile.credentials[1].uri | string | `"repo1.dso.mil"` |  |
| credentialsFile.credentials[1].username | string | `""` |  |
| credentialsFile.credentials[1].password | string | `""` |  |
| baseConfig.bbctl-log-add-source | bool | `true` |  |
| baseConfig.bbctl-log-format | string | `"json"` |  |
| baseConfig.bbctl-log-level | string | `"warn"` |  |
| baseConfig.bbctl-log-output | string | `"stderr"` |  |
| baseConfig.big-bang-repo | string | `"https://repo1.dso.mil/big-bang/bigbang/-/blob/master/"` |  |
| baseConfig.big-bang-skip-auto-update | bool | `true` |  |
| baseConfig.output-config.format | string | `"json"` |  |
| baseConfig.output-config.json-pretty-print | bool | `false` |  |
| baseConfig.skip-update-check | bool | `true` |  |
| baseConfig.util-credential-helper.big-bang-credential-helper-credentials-file-path | string | `"/home/bigbang/.bbctl/credentials.yaml"` |  |
| baseConfig.util-credential-helper.big-bang-credential-helper | string | `"credentials-file"` |  |
| baseConfig.preflight-check.registryserver | string | `""` |  |
| baseConfig.preflight-check.registryusername | string | `""` |  |
| baseConfig.preflight-check.registrypassword | string | `""` |  |
| baseLabels | object | `{}` |  |
| bigbangUpdater.enabled | bool | `true` |  |
| bigbangUpdater.importDashboards | bool | `true` |  |
| bigbangUpdater.schedule | string | `"0 * * * *"` |  |
| bigbangUpdater.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangUpdater.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangUpdater.labels | object | `{}` |  |
| bigbangUpdater.config | object | `{}` |  |
| bigbangUpdater.podAnnotations | object | `{}` |  |
| bigbangUpdater.serviceAccount.create | bool | `true` |  |
| bigbangUpdater.serviceAccount.annotations | object | `{}` |  |
| bigbangUpdater.serviceAccount.name | string | `""` |  |
| bigbangStatus.enabled | bool | `true` |  |
| bigbangStatus.importDashboards | bool | `true` |  |
| bigbangStatus.schedule | string | `"0 * * * *"` |  |
| bigbangStatus.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangStatus.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangStatus.labels | object | `{}` |  |
| bigbangStatus.config | object | `{}` |  |
| bigbangStatus.podAnnotations | object | `{}` |  |
| bigbangStatus.serviceAccount.create | bool | `true` |  |
| bigbangStatus.serviceAccount.annotations | object | `{}` |  |
| bigbangStatus.serviceAccount.name | string | `""` |  |
| bigbangViolations.enabled | bool | `true` |  |
| bigbangViolations.importDashboards | bool | `true` |  |
| bigbangViolations.schedule | string | `"0 * * * *"` |  |
| bigbangViolations.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangViolations.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangViolations.labels | object | `{}` |  |
| bigbangViolations.config | object | `{}` |  |
| bigbangViolations.podAnnotations | object | `{}` |  |
| bigbangViolations.serviceAccount.create | bool | `true` |  |
| bigbangViolations.serviceAccount.annotations | object | `{}` |  |
| bigbangViolations.serviceAccount.name | string | `""` |  |
| bigbangPreflight.enabled | bool | `true` |  |
| bigbangPreflight.importDashboards | bool | `true` |  |
| bigbangPreflight.schedule | string | `"0 * * * *"` |  |
| bigbangPreflight.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangPreflight.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangPreflight.labels | object | `{}` |  |
| bigbangPreflight.config | object | `{}` |  |
| bigbangPreflight.podAnnotations | object | `{}` |  |
| bigbangPreflight.serviceAccount.create | bool | `true` |  |
| bigbangPreflight.serviceAccount.annotations | object | `{}` |  |
| bigbangPreflight.serviceAccount.name | string | `""` |  |
| bigbangPolicy.enabled | bool | `true` |  |
| bigbangPolicy.importDashboards | bool | `true` |  |
| bigbangPolicy.policyEnforcer | string | `"kyverno"` |  |
| bigbangPolicy.schedule | string | `"0 * * * *"` |  |
| bigbangPolicy.bigbangReleaseName | string | `"bigbang"` |  |
| bigbangPolicy.bigbangReleaseNamespace | string | `"bigbang"` |  |
| bigbangPolicy.labels | object | `{}` |  |
| bigbangPolicy.config | object | `{}` |  |
| bigbangPolicy.podAnnotations | object | `{}` |  |
| bigbangPolicy.serviceAccount.create | bool | `true` |  |
| bigbangPolicy.serviceAccount.annotations | object | `{}` |  |
| bigbangPolicy.serviceAccount.name | string | `""` |  |

## Contributing

Please see the [contributing guide](./CONTRIBUTING.md) if you are interested in contributing.

---

_This file is programatically generated using `helm-docs` and some BigBang-specific templates. The `gluon` repository has [instructions for regenerating package READMEs](https://repo1.dso.mil/big-bang/product/packages/gluon/-/blob/master/docs/bb-package-readme.md)._

