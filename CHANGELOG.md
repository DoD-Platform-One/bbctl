# Changelog

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [3.0.1-bb.1] (2026-1-2)
### Changed
- updated bbctl to application version 2.2.0
- gluon updated from 0.9.5 to 0.9.7
- updated registry1.dso.mil/ironbank/opensource/yq/yq (source) 4.48.1 -> 4.50.1
- updated registry1.dso.mil/ironbank/redhat/ubi/ubi8-minimal (source) 9.6 -> 9.7

## [3.0.1-bb.0] (2025-10-15)
### Changed
- updated bbctl to application version 2.1.0
- gluon updated from 0.9.0 to 0.9.5
- updated registry1.dso.mil/ironbank/opensource/yq/yq (source) 4.47.2 -> 4.48.1

## [3.0.0-bb.0] (2025-09-15)
### Changed
- updated bbctl to application version 2.0.0
- added new registry-override section to the app config to support custom container registries in multiple commands
- added preflight check image override value to support custom image (defaults to registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.6 when empty)
- removed preflight check credentials from app config
- gluon updated from 0.8.0 to 0.9.0
- updated registry1.dso.mil/ironbank/opensource/yq/yq (source) 4.47.1 -> 4.47.2

## [2.1.0-bb.2] (2025-08-16)
### Changed
- gluon updated from 0.7.0 to 0.8.0

## [2.1.0-bb.1] (2025-07-24)
### Changed
- gluon updated from 0.6.3 to 0.7.0
- updated registry1.dso.mil/ironbank/opensource/yq/yq (source) 4.45.4 -> 4.47.1
- updated bbctl to application version 1.5.0
- updated ubi-minimal image major version from 8 to 9
  - registry1.dso.mil/ironbank/redhat/ubi/ubi8-minimal:8.10 -> registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.6

## [2.1.0-bb.0] (2025-07-09)
### Changed
- Modified chart to use a secret for credentials instead of a configmap.

## [2.0.0-bb.3] (2025-06-27)
### Changed
- gluon updated from 0.6.2 to 0.6.3
- Upgraded bbctl to application version 1.4.0

## [2.0.0-bb.2] (2025-06-18)
### Changed
- set "concurrencyPolicy: Forbid"

## [2.0.0-bb.1] (2025-06-04)
### Changed
- adjusted cypress test to use gluon grafana logon built in command

## [2.0.0-bb.0] (2025-06-02)
### Changed
- updated all dashboard queries label selector from "app" to "app_kubernetes_io_name" to work with Alloy
### Deprecated
- Promtail is now deprecated and is only supported by the 1.0.0-bb.x chart versions

## [1.0.0-bb.11] (2025-06-02)
### Changed
- gluon updated from 0.5.19 to 0.6.2

## [1.0.0-bb.10] (2025-05-28)
### Changed
- fix typo error in cypress test

## [1.0.0-bb.9] (2025-05-22)
### Changed
- cypress tests to account for Grafana v12 UI updates

## [1.0.0-bb.8] (2025-05-15)
### Changed
- gluon updated from 0.5.15 to 0.5.19
- updated registry1.dso.mil/ironbank/opensource/yq/yq (source) 4.45.1 -> 4.45.4
- updated registry1.dso.mil/ironbank/redhat/ubi/ubi8-minimal (source) 8.4 -> 8.10

## [1.0.0-bb.7] (2025-05-13)

### Changed
- Added helm images annotation to Chart.yaml
- Updated resource limits to reduce OOM issues

## [1.0.0-bb.6] (2025-05-07)

### Changed

- Upgraded bbctl to application version 1.2.0
- Added bbctl base configuration values for formatting output
- Added bbctl base configuration values for skipping automatic big bang repository checkout updates
- Added bbctl base configuration values for skipping automatic bbctl update checks

## [1.0.0-bb.5] (2025-04-22)

### Changed

- Added "bigbang.dev/applicationVersions" annotation to the chart

## [1.0.0-bb.4] (2025-04-10)

### Changed

- gluon - patch - 0.5.14 -> 0.5.15

## [1.0.0-bb.3] - 2025-04-08

### Changed

- Enable the standard pipelines for packages

## [1.0.0-bb.2] - 2025-03-20

### Changed

- Added Istio custom authorization policies template to work with Istio hardening configurations
- Updated README to include the new istio field value descriptions

## [1.0.0-bb.1] - 2025-03-13

### Changed

- Added Istio custom service entry template to work with Istio hardening configurations
- Updated structure of Big Bang values in values.yaml

## [1.0.0-bb.0] - 2025-03-04

### Changed

- Upgraded bbctl to application version 1.0.0

## [0.7.6-bb.2] - 2025-01-29

### Changed

- cypress - major - ^13.0.0 -> ^14.0.0
- gluon - patch - 0.5.12 -> 0.5.14
- Updated renovate.json to work for independent chart repo structure

## [0.7.6-bb.1] - 2025-01-06

### Changed

- Updated chart to use latest bbctl 0.7.6 version

## [0.7.6] - 2025-01-03

### Changed

- Updated golang/x/crypto upstream dependency

## [0.7.5-bb.0] - 2024-11-22

### Added

- Added the maintenance track annotation and badge

## [0.7.5] - 2024-11-06

### Changed

- Security updates
- Added init command
- Added set command
- Finished standardizing logging and regular output
- Added code coverage minimums
- Switched to a new yaml library
- Updated golang
- Bug fixes

## [0.7.4-bb.0] - 2024-09-16

### Added

- Added helm chart for bbctl
- Updated the release process to include helm chart

## [0.7.4] - 2024-09-16

### Changed

- Security updates
- Updated the Makefile to include more commands
- Added more documentation
- Bubbled up errors to standardize error handling and remove panics
- Started standardizing logging and regular output
- Added version update detection functionality
- Bug fixes

## [0.7.3] - 2024-07-11

### Changed

- Vetted and extended testing to >=80% coverage
- Added more documentation
- Standardized help output
- Updated the Makefile to include more commands

## [0.7.2] - 2024-05-20

### Changed

- Preparing everything for kickoff presentation

## [0.7.1] - 2024-04-19

### Changed

- Fixing some tests that fail in the post pipeline

## [0.7.0] - 2024-04-19

### Changed

- Added deploy and k3d commands
- linted
- formatted
- refactored into more factories
- added a lot of testing
- added the Makefile

## [0.6.0] - 2023-10-24

### Changed

- Update Contributing Guidelines

## [0.6.0] - 2022-08-01

### Changed

- upgrade package dependencies

## [0.5.0] - 2022-07-18

### Changed

- Add kyverno policy violations

## [0.4.0] - 2021-03-04

### Changed

- upgrade package dependencies

## [0.3.0] - 2021-02-28

### Changed

- add research help to output of status command

## [0.2.1] - 2021-02-18

### Changed

- move pipeline code to CI configuration

## [0.2.0] - 2021-02-08

### Added

- create MVP
