# bbctl Development and Maintenance Guide

By default, this chart is configured to utilize the `registryCredentials` passed in from the Big Bang umbrella chart to create a secret for image pulls.

## Big Bang `test-values.yaml` Configuration

The following is a minimal example for enabling `bbctl` within a Big Bang deployment `test-values.yaml`. Note that `bbctl` is enabled by default in Big Bang. This chart is dependent on `monitoring` and `loki` for full functionality.

```yaml
bbctl:
  enabled: true

monitoring:
  enabled: true

loki:
  enabled: true
```

## Updating The Chart

1.  Update the `version` in `./chart/Chart.yaml`. The `appVersion` should automatically be updated by renovate as new releases are made.

2.  Update `CHANGELOG.md` with an entry for the new version, noting all changes.

3. Regenerate the `README.md` file.

## Unit Testing Helm Templates

This chart uses the [`helm-unittest`](https://github.com/helm-unittest/helm-unittest) plugin to enable unit testing of Helm templates.

### Installation

Install the plugin via the following command:

```shell
helm plugin install https://github.com/helm-unittest/helm-unittest.git
```

### Running Tests

To run the test suite, navigate to the chart's root directory and execute the plugin:

```shell
cd chart
helm unittest .
```

## Manually Triggering CronJobs

The bbctl data is populated by running CronJobs. These may not be triggered for up to an hour by default. For testing purposes, it may be useful to manually trigger the jobs immediately.

To manually trigger the cronjobs for testing purposes, you can use the following commands:

```shell
kubectl create job --from=cronjob/bbctl-bbctl-bigbang-policy bbctl-bbctl-bigbang-policy-manual -n bbctl
kubectl create job --from=cronjob/bbctl-bbctl-bigbang-preflight bbctl-bbctl-bigbang-preflight-manual -n bbctl
kubectl create job --from=cronjob/bbctl-bbctl-bigbang-status bbctl-bbctl-bigbang-status-manual -n bbctl
kubectl create job --from=cronjob/bbctl-bbctl-bigbang-updater bbctl-bbctl-bigbang-updater-manual -n bbctl
kubectl create job --from=cronjob/bbctl-bbctl-bigbang-violations bbctl-bbctl-bigbang-violations-manual -n bbctl
```
