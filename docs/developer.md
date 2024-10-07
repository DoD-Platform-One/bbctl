# bbctl

CLI tool to simplify development, deployment, auditing and troubleshooting of Big Bang in a kubernetes cluster.

## Contributing

Code contributions from the community are welcomed. Steps to contribute:
1. Create an issue in this project. See the [issues page](https://repo1.dso.mil/big-bang/product/packages/bbctl/-/issues)
1. Fill in relevant information about the issue so that others can understand what it is for 
1. Assign yourself to the issue so that it is clear that you are contributing code verses just reporting an issue.
1. View the issue in the Gitlab UI, and from there you can create a branch and a corresponding merge request.
1. A simple pipeline pipeline with linting and unit tests will automatically run for merge requests.
1. When your code is ready add a ```status::review``` label to the merge request
1. Code owners will review, test, and merge as appropriate.
1. Code owners will create a release tag and a package will be built by the PartyBus mission devops pipeline.

### Contribution conditions

1. The code must include a minimum of 80% unit test coverage
1. The code must pass lint test
1. Help resolve any security issues found in the mission ops pipeline
1. Commands should be well documented and adhere to the guidelines in the [Command Guidelines](./bbctl-command-guidelines.md)

## Development Environment 

The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

This project supports [devcontainers](https://containers.dev/), [devpod](https://devpod.sh/docs/getting-started/install), and [devbox](https://www.jetify.com/devbox/) all of which will instantly give you an environment to work in. It also uses a Makefile to provide helpful development scripts.

__NOTE:__ These folders have to exist on your machine for the container to start, they don't have to have anything in them, but they do have to be there.
- `~/.ssh`
- `~/.kube`
- `~/.bbctl`
- `~/.aws`
- `~/.gitconfig`
- `~/.config`

### Dev Containers

For the simplest most integrated dev environment in VS Code, install the [Dev Containers Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers). Then every time you open the workspace it will offer to reopen the workspace in the dev container.
- `F1` then type `reopen` and you can easily switch between developing locally and in a container

### Devpod

To run with devpod you can use the Makefile:
- `make dup` will start the container (and build it if necessary)
- `make dstop` will stop the container
- `make dbuild` will recreate the container and start it

This solution also let you select kubernetes as a provider meaning your shell and vscode server can live inside your working cluster.

### Devbox

Both of the dev container based solutions use devbox, just inside a container. Here are some basic commands for interacting with it (in or outside of a container):
- `devbox run [script name]` - runs the named script in the devbox.json
- `devbox shell` - Start a shell in the devbox instance
- `devbox install` - Installs all of the packages in the devbox.json
- `devbox update` - Updates the packages in devbox.json
- `devbox search` - Search for nix packages
- `devbox add` - Add a package to the devbox.json
- `devbox help` - Get help overview
- `devbox [command] --help` - Get help for a command

### Install Golang

Follow the instructions in official Go document for the specific development platform:
https://golang.org/doc/install

Define an environment variables GOPATH and GOROOT
```bash
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
```

Create directories and set PATH environment variable
```bash
mkdir -p $HOME/go/{bin,src,pkg}
export PATH="$PATH:${GOPATH}/bin"
```

Clone the repo such that the `bbctl` is available in the following location (clone it there or run `ln -s real-location gopath-location`):
```bash
$GOPATH/src/repo1.dso.mil/big-bang/product/packages/bbctl
```

Make the environment variables permanent by setting them in your shell's rc `~/.bash_rc` or equivalent for alternative shells.
```bash
# support for GoLang development
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH="$PATH:${GOPATH}/bin:${GOROOT}/bin"
```

### Install cobra

```bash
go get -u github.com/spf13/cobra
```

### Add new commands with cobra

The base command is defined in `cmd.go` and new subcommands are added in `NewRootCmd` function. Follow `list.go` as an example to create a new subcommand. Refer to [command semantics](/docs/command.md) for the practices followed in naming and implementing `bbctl` commands.

### Build only with no local install

Execute the following from the project root to build the binary without local install
```bash
make build
# OR
go build
```

Run the built binary using dot-slash
```bash
make run version
# OR
./bbctl version
```

### Build and Install

Execute the following from the project root to build the executable and make it available in $GOPATH/bin directory:
```bash
make install
# OR
go install
```

Run the installed `bbctl` tool
```bash
bbctl version
```

### Run unit tests

```bash
make coverage
# OR
make test
# OR
go test -v ./... -coverprofile=cover.txt
```

### Run lint checks

Linting checks code quality based on best practice. For now the [linter tool](https://golangci-lint.run/welcome/install/) is no longer [the one from the golang project](https://github.com/golang/lint) as it's deprecated. To manually run the linter follow these steps.  
1. install the tool
    ```bash
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    ```
2. Run the linter from this project's root directory
    ```bash
    make lint
    # OR
    golangci-lint run ./...
    ```
3. Set it up in vscode
    1. Open User Settings (visual)
    1. Search for golang
    1. Change the `Go: Lint Tool` to `golangci-lint`

## Development Tasks

Here some common development tasks will be laid out with common issues and solutions.

### Upgrading

```bash
go get -u

# You should immediately build and run tests afterwards
make all
# OR
go build
go test -v -coverprofile=test.out -cover ./...
```

#### Problem Packages

1. `oras.land/oras-go` and `github.com/docker/docker`
    1. If you get failures from doing that in relation to `oras.land/oras-go` check if the package `github.com/docker/docker` got upgraded. Oras seems to depend on a bunch of incompatible versions of packages and that one causes build issues.
    1. Last seen 1/19/2024

#### Debugging New Problem Packages

1. If that isn't the issue use the following to find it's dependencies.
    ```bash
    go mod graph | grep "name/of/package/throwing/error.go"
    ```
1. Revert all of those, then see if it works with a build/test.
    1. If it does commit, then start adding those upgrades back in one at a time and make a list of problem package(s)
    1. If it doesn't
        1. If there is a new error start this process again for that error as well
        1. If it's the same error, ensure you reverted all of the dependencies
            1. Most often one was simply missed from the list generated by the initial `go mod graph`
            1. Note some may be intermediate meaning you'd need to do the `go mod graph` for the level 1 dependencies to get the level 2. This is relatively rare though.
1. Note the new problem packages [here](#problem-packages)

## Helm Chart

The `bbctl` tool is also available as a helm chart in `chart/`. This chart is used to deploy the `bbctl` tool in a kubernetes cluster. The chart is built using the [helm](https://helm.sh/) package manager for kubernetes. The tool is deployed as a set of cronjobs in the cluster and runs periodically to perform various tasks.

### Big Bang Integration

`bbctl` is not currently integrated with Big Bang. It can be deployed as a `.package` using the values below. It will eventually be integrated into the Big Bang chart.

```yaml
kyvernoPolicies:
  values:
    # validationFailureAction: "audit"
    policies:
      disallow-auto-mount-service-account-token:
        exclude:
          any:
          - resources:
              namespaces:
              - "bbctl"
              kinds:
              - "Pod"
              - "CronJob"
              - "Job"
              - "ServiceAccount"
              names:
              - "bbctl*"

istio:
  enabled: true

monitoring:
  enabled: true

packages:
  bbctl:
    enabled: true
    sourceType: "git"
    git:
      repo: https://repo1.dso.mil/big-bang/product/packages/bbctl.git
      path: chart
      tag: 0.7.4-bb.0
    #   tag: null
    #   branch: 197-implement-helm-chart
    flux:
      timeout: 5m
    postRenderers: []
    dependsOn:
      - name: monitoring
        namespace: bigbang
      - name: kyverno-policies
        namespace: bigbang
    wrapper:
      enabled: true
    network:
      allowControlPlaneEgress: true
    values: {}
```

### Versioning and Release

The CHANGELOG versions will include both the version of the chart and the version of the `bbctl` tool. The version of the tool will follow [semver](https://semver.org/) `x.x.x` and the version of the chart will follow the format `x.x.x-bb.x`. The version of the chart will be incremented with each release of the chart. The version of the tool will be incremented with each release of the tool. The chart version will be reset to `x.y.z-bb.0` with each new version of the tool versioned `x.y.z`.

As part of the validation of MRs the chart's `.appVersion`, the image tag, the most recent CHANGELOG app version, and `bbctl version --client | awk '{print $4}'` will all match; in addition, the chart's `.version` will be the same as the most recent CHANGELOG chart version. The chart will be tested in a cluster to ensure it is working as expected. The tool will be tested with unit and integration tests to ensure it is working as expected. The versions will be different than the version on main (should be some kind of incrementing). Once merged a gitlab release will be created and the chart or app will be available for use. As part of the tool release process, the [docker image](https://repo1.dso.mil/dsop/big-bang/bbctl) will get an MR opened with the new versions and artifacts. Once that MR is merged, the image is built and pushed to the registry. This will make the chart upgradable.

Once this is integrated into Big Bang, after the gitlab release is created the Big Bang MR to update the chart will be created. This is the last step in the release process for `bbctl`.
