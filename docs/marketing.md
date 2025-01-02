# bbctl

`bbctl` is a command line tool designed to help you deploy, manage, and upgrade your Big Bang clusters. It is intended to be your one-stop-shop for all things Big Bang and assist in preparing a cluster for Big Bang, deploying Big Bang into the cluster, and maintaining and utilizing your cluster.

## CLI Installation
`bbctl` as a CLI can be installed by the following the [installation instructions](https://repo1.dso.mil/big-bang/product/packages/bbctl/-/blob/main/docs/user-guide.md#installation) in the `bbctl` repository. After installing the binary, `bbctl config init` can be run to start a guided configuration wizard.

## Dashboard Installation

In addition to the CLI commands, `bbctl` can be deployed into the cluster as a Big Bang addon. When the helm chart is installed, it can be configured to run commands as CronJobs and report command the results periodically.
- Note: this functionality is not 100% implemented yet and the docs and process are being refined in: https://repo1.dso.mil/big-bang/product/packages/bbctl/-/issues/199. 

`bbctl` will also create a collection of dashboards in the Big Bang's Grafana instance (if installed) for easier consumption of results. These can provide additional visibility into the health of the cluster and the Big Bang deployment.

## Key Features

### Prepare Kubernetes Clusters for Big Bang
Cluster admins can utilize the `bbctl preflight-check` command to verify that their cluster is prepared to run Big Bang. Preflight checks are run against the cluster and validate various components of the cluster are in place and ready to be used, such as flux, metrics server, and various configurations like compatible storage checks. The preflight check prints out a simple report with the current status of dependencies as well as remediation steps if errors occur.

### Deploy Big Bang clusters
`bbctl` aims to simplify the process of deploying Big Bang into a cluster. Rather than running a long, complicated `helm` command, `bbctl` providers a simpler `bbctl deploy` interface that removes a lot of the boilerplate.
-  `bbctl deploy flux`
    - This command can load registry credentials from a file or a password manager and deploy `flux` into your cluster pre-configured.
- `bbctl deploy bigbang`
    - This command removes a lot of the complexity of crafting long `helm` commands and can deploy Big Bang into your cluster with a single command.
    - Optional addon packages can easily be enabled with the `--addons` flags and additional configuration can be passed to the underlying `helm install` command for even further control.

### Manage Big Bang clusters
There are a variety of commands in `bbctl` that help SREs manage and administer their Big Bang clusters:

- `bbctl version`
    - The version command shows the current version of `bbctl`, `bigbang`, and all installed addons.
    - Optionally, with `--check-for-updates`, `bbctl` will retrieve the most up-to-date version of the requested services and inform the user if an update is available

- `bbctl policy` 
    - The policy command can inform the user about currently configured Kyverno and Gatekepper policies and inspect their configurations, and `bbctl violations` will list all policy violations from the installed policy management engine.

- `bbctl status` 
    - The status command will show the health status of the installed compoents of Big Bang and the cluster

- `bbctl help`
    - The help command will output the complete list of commands avaialable in `bbctl`.
    - Additional usage information for each command can be accessed by adding the `-h` flag.

## Future Work
`bbctl` is under active development and will continue to be improved. The core functionality is currenty implemented but the team is working hard to refine the user experience and add additional features prior to the 1.0 release. Still in the pipeline is:
- [Integrate Repository to BB](https://repo1.dso.mil/big-bang/product/packages/bbctl/-/issues/199) - This ticket will fully integrate `bbctl` into the Big Bang produt and enable deployment as easily as any other package
- [Add periodic Self-Update Feauture](https://repo1.dso.mil/groups/big-bang/-/epics/304) - As the name suggests, we are adding in self-updating functionality to `bbctl` to allow for easy updates to the CLI
- [Create bbctl Config File Versioning Process](https://repo1.dso.mil/groups/big-bang/-/epics/396) - This epic is focused on creating internal checks that local `bbctl` configurations are up to date with the current deployed version of the tool


