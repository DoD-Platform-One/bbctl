# User Guide
The BBCTL Command line interface(CLI) tool has been created to simplify development, deployment, auditing and troubleshooting of the BigBang product in a kubernetes cluster. The bbctl repository is mirrored to PartyBus code.il2.dso.mil where a Mission DevOps pipelne is run and a package is built and pushed back to repo1.dso.mil. The code has passed security scans and is eligible to recieve a certificate to field(CTF). 

## Installation
1. Navigate to the [Package Registry Page](https://repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/-/packages). 
1. Locate the most current package based on semanic versioning. Click on the package name. 
1. Download the package by clicking on the ```bbctl-x.x.x.tar.gz``` archive.
1. Extract the archive
    ```bash
    tar xvzf ~/Downloads/ bbctl-.x.x.x.tar.gz
    ```
1. There are binaries built for Linux and Mac. Based on your operating system move the appropriate binary to a directory that is included in your workstation path. Typically ```/usr/local/bin/```  
    Linux
    ```bash
    # move the downloaded binary
    sudo mv ~/Downloads/bbctl-linux-amd64 /usr/local/bin/
    # create symbolic link
    sudo ln -s /usr/local/bin/bbctl-linux-amd64 /usr/local/bin/bbctl
    # test the version
    bbctl version
    ```
    Mac
    ```bash
    # move the downloaded binary
    sudo mv ~/Downloads/bbctl-darwin-amd64 /usr/local/bin/
    # create symbolic link
    sudo ln -s /usr/local/bin/bbctl-darwin-amd64 /usr/local/bin/bbctl
    # test the version
    bbctl version
    ```

## Usage
The bbctl tool is self documenting so only a few simple examples are included here. The bbctl commands work similar to other well known tools such as ```kubectl```
```
# get help for commands
bbctl -h
# get bbctl version
bbctl version
# preflight check: Checks status of k8s cluster before deploying BigBang
bbctl preflight-check --registryserver https://registry1.dso.mil --registryusername your.name --registrypassword yourPassword
# git status of BigBang deployment
bbctl status
# get the helm chart values for a helm release as deployed by BigBang
bbctl values RELEASE_NAME
```

## Command completion
To enable command completion using the tab key, ensure that bbctl completion script gets sourced in all your shell sessions. Execute the following command for details on how to generate the completion script and load it in the supported shells:
```
bbctl completion -h
```
