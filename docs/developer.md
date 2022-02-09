# bbctl
CLI tool to simplify development, deployment, auditing and troubleshooting of BigBang in a kubernetes cluster.

## Contributing
Code contributions from the community are welcomed. Steps to contribute:
1. Create an issue in this project. See the [issues page](https://repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/-/issues)
1. Fill in relevant information about the issue so that others can understand what it is for 
1. Assign yourself to the issue so that it is clear that you are contributing code verses just reporting an issue.
1. View the issue in the Gitlab UI, and from there you can create a branch and a corresponding merge request.
1. A simple pipeline pipeline with linting and unit tests will automatically run for merge requets.
1. When your code is ready add a ```status::review``` label to the merge request
1. Code owners will review, test, and merge as appropriate.
1. Code owners will create a release tag and a package will be built by the PartyBus mission devops pipeline.
### Contribution conditions
1. The code must include a minimum of 80% unit test coverage
1. The code must pass lint test
1. Help resolve any security issues found in the mission ops pipeline

## Development Environment 
The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

### Install Golang
Follow the instructions in official Go document for the specific development platform:
https://golang.org/doc/install

Define an environment variables GOPATH and GOROOT
```
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
```
Create directories and set PATH environment variable
```
mkdir -p $HOME/go/{bin,src,pkg}
export PATH="$PATH:${GOPATH}/bin"
```
Clone the repo such that the bbctl is available in the following location:
```
$GOPATH/src/repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl
```
Make the environment variables permanent by setting them in your shell profile   
~/.bash_profile
```
# support for GoLang development
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH="$PATH:${GOPATH}/bin:${GOROOT}/bin"
```
You might also have to source the bash profile in the bashrc
~./bashrc
```
source ~/.bash_profile
```

### Install cobra
```
go get -u github.com/spf13/cobra/cobra
```

### Add new commands with cobra
The base command is defined in cmd.go and new subcommands are added in NewRootCmd function. Follow list.go as an example to create a new subommand. Refer to [command semantics](./docs/command.md) for the practices followed in naming bbctl commands.

### Build only with no local install
Execute the following from the project root to build the binary without local install
```
go build
```
Run the built binary using dot-slash
```
./bbctl -h
```

### Build and Install
Execute the following from the project root to build the executable and make it available in $GOPATH/bin directory:
```
go install
```
Run the installed bbctl tool
```
bbctl -h
```

### Run unit tests
```
go test -v ./... -coverprofile=cover.txt
```

### Run lint checks
Linting checks code quality based on best practice. For now the [linter tool](https://github.com/golang/lint) is the one from the golang project. To manually run the linter follow these steps.  
1. install the tool
    ```
    go install golang.org/x/lint/golint@latest
    ```
2. Run the linter from this project's root directory
    ```
    golint -set_exit_status ./...
    ```


