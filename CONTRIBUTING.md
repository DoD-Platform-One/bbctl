# Contributing

Thanks for contributing to this repository!

If you are coming from `repo1.dso.mil` and have an account at `login.dso.mil` please keep reading. If you are coming from or looking for the [project on Github](https://github.com/DoD-Platform-One) and wanting to make a Pull Request without a `dso.mil` account please see [External Github Contributions](https://repo1.dso.mil/big-bang/bigbang/-/blob/master/CONTRIBUTING.md?ref_type=heads#community-contributions-to-dod-platform-one-via-github).

This repository follows the following conventions:

* [Semantic Versioning](https://semver.org/)
* [Keep a Changelog](https://keepachangelog.com/)
* [Conventional Commits](https://www.conventionalcommits.org/)

Development requires the Kubernetes CLI tool as well as a local Kubernetes cluster. [KIND](https://github.com/kubernetes-sigs/kind) is recommended as a lightweight local option for standing up Kubernetes clusters.

To contribute a change:

1. Create a branch on the cloned repository with a descriptive name. It helps with project tracking when the branch name starts with the Gitlab issue number,for example: 7-issue-description
1. Make the changes in code.
1. Write [GoLang unit tests](https://golang.org/doc/tutorial/add-a-test) as appropriate.
1. Write [Cypress tests](https://docs-bigbang.dso.mil/latest/docs/guides/using-bigbang/testing-deployments/#default-configuration-options) as appropriate and ensure they pass when [running them within a cluster](https://repo1.dso.mil/big-bang/product/packages/gluon/-/blob/master/docs/executing-cypress.md) with Big Bang deployed.
1. Make commits using the [Conventional Commits](https://www.conventionalcommits.org/) format. This helps with automation for changelog. Update `CHANGELOG.md` in the same commit using the [Keep a Changelog](https://keepachangelog.com). Depending on tooling maturity, this step may be automated.
1. Open a merge request using one of the provided templates. If this merge request is solving a preexisting issue, add the issue reference into the description of the MR.
1. During this time, ensure that all new commits are rebased into your branch so that it remains up to date with the `main` branch.
1. Wait for a maintainer of the repository (see CODEOWNERS) to approve.
1. If you have permissions to merge, you are responsible for merging. Otherwise, a CODEOWNER will merge the commit.
