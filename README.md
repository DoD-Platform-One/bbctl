# bbctl
CLI tool to simplify development, deployment, auditing and troubleshooting of BigBang in a kubernetes cluster.

## Development Environment 
The CLI tool is developed in Go language and uses the [cobra](https://github.com/spf13/cobra/) library to implement commands.

### Install Golang
Follow the instructions in official Go document for the specific development platform:
https://golang.org/doc/install

Define an environment variables GOPATH and GOROOT
```
export GOPATH=$HOME/go/
export GOROOT=/usr/local/go/
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
export GOPATH=$HOME/go/
export GOROOT=/usr/local/go/
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
The base command is defined in cmd.go and new subcommands are added in NewRootCmd function. Follow list.go as an example to create a new subommand.

### Build and Install
Execute the following to build the executable and make it available in $GOPATH/bin directory:
```
go install
```

### Execute bbctl
```
bbctl -h
```

### Run unit tests
```
go test -v ./...
```

### Run lint checks
Linting checks code quality based on best practice. For now the [linter tool](https://github.com/golang/lint) is the one from the golang project. To manually run the linter follow these steps.  
1. install the tool
    ```
    go get -u golang.org/x/lint/golint
    ```
2. Undo any modifications the install made to the go.mod file
3. Run the linter from this project's root directory
    ```
    golint -set_exit_status ./...
    ```

### Command completion

To enable command completion using the tab key, ensure that bbctl completion script gets sourced in all your shell sessions. Execute the following command for details on how to generate the completion script and load it in the supported shells:
```
bbctl completion -h
```

