# BBCTL

Command line interface(CLI) tool to simplify development, deployment, auditing and troubleshooting of the BigBang product in a kubernetes cluster.

## User Guide

Follow the [user guide](/docs/user-guide.md) for how to install and use the bbctl tool.

## Developer Documentation

Help Contribute! See the [developer documentation](/docs/developer.md). The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

## `bbctl` Usage and Design Priorities

### Automated usage over interactive usage

`bbctl` is primarily intended to be piped to/from other tools and shell scripts. Interactive use is a future possibility.

### Multiple execution contexts

`bbctl` will be running as a cronjob in cluster, possibly as web server in cluster, potentially in pipelines, and on developer machines.

### External _and_ internal users

`bbctl` is currently used both inside and outside the BigBang team as a fully open source project.
