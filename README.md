# bbctl

CLI tool to simplify development, deployment, auditing and troubleshooting of BigBang in a kubernetes cluster.

## Development Environment 

The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

### Install Golang

Follow the instructions in official Go document for the specific development platform:

https://golang.org/doc/install

Define an environment variable GOPATH 

```$ export GOPATH=$HOME/go```

Create directories

```$ mkdir -p $HOME/go/{bin,src,pkg}```

```$ export PATH="$PATH:${GOPATH}/bin"```

Clone the repo such that the bbctl is available in the following location:

```$GOPATH/src/repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl```

### Install cobra

```$ go get -u github.com/spf13/cobra/cobra```

### Add new commands with cobra

The base command is defined in cmd.go and new subcommands are added in NewRootCmd function. Follow list.go as an example to
create a new subommand.

### Build bbctl

Execute the following to build the executable and make it available in $GOPATH/bin directory:

```$ go install```

### Execute bbctl

```$ bbctl -h```

### Run unit tests

```$ go test -v ./...```
