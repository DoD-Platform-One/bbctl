# Completion Tests
This scripts help create consistent testing environments of various shell environments for usage in testing `bbctl completion`.

## Usage
`./test-shell.sh -s <shell of your choice>`

This will start an epehermal docker container with a minimal installation of `bbctl` preconfigured with your current `$KUBECONFIG` and `big-bang-repo` configurations. From here, you should be able to run `bbctl completion` and generate completions for the correct shell. Use this for interactive testing.
