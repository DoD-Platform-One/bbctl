# `bbctl` Command Usage Guidelines

This document is a guide to writing usage strings for `bbctl` subcommands. This is a living document and will be updated as the project evolves. Contributions are welcome.

The hope of this document is to create a set of guidelines useful for writing and auditing usage strings for `bbctl` subcommands. By defining these guidelines in a public place, we can hopefully reduce the variance in usage strings across the `bbctl` subcommands and create a more consistent experience for users.

`bbctl` aims to feel familiar and intuitive to users of `kubectl` and `helm`. When designing commands, we should aim to follow similar conventions and patterns as those teams. If the invocation of a command feels awkward in a cloud-native environment, we should aim to make it feel more familiar.

## Usage Strings

All commands should implement the following strings:
- Usage String - the one-line usage message
- Short Description - the short description to show in the `--help` output
- Long Description - the long description to show in the `--help` output
- Example - a list of examples of how to use the command

### Annotated Example (from `cmd/violations.go`)
Below is a annotated example of a command that implements the above values. Descriptions and expectations are provided in the comments.

In general, we should err on the side of being too verbose in our documentation and usage strings. Provide as many varied examples as possible, and capture any details that are likely to be useful to users in the long description.

```go

var (
    // violationsUse is the usage string for the command. This should be a one-line usage message that encapsulates the default, minimal
    // invocation of the command.
	violationsUse = `violations`

    // Short, one line description of the the given command.
	violationsShort = i18n.T(`List policy violations.`)

    // Long, multi-line description of the the given command. This is printed when `--help` is used. The long description should
    // be a complete description of the command and should include any required context and commentary about how to correctly use the configuration.
    // e.g. This description should note whether it's dependent on a given configuration or addon to be enabled in the cluster.
    // Long descriptinos should err on the side of being too verbose and not too terse. This will be the primary way users will 
    // attempt to debug unexpected behaviors. violationsLong = templates.LongDesc(i18n.T(`
		List policy violations reported by Gatekeeper or Kyverno Policy Engine.

		Note: In case of kyverno, violations are reported using the default namespace for kyverno policy resource
		of kind ClusterPolicy irrespective of the namespace of the resource that failed the policy. Any violations
		that occur because of namespace specific policy i.e. kind Policy is reported using the namespace the resource
		is associated with. If it is desired to see the violations because of ClusterPolicy objects, use the command
		as follows:

		bbctl violations -n default
	`))

    // List many examples of how to use the command. Examples should include a simple bash-comment above them explaining
    // the end result of the example. Try to include a variety of examples on commands with complicated invocations or many flags.
    // Also try and capture the most-likely use cases for the command first, and increase complexity of the examples as needed.
    // Variables passed to commands should be included in the examples in all caps, like NAMESPACE.
	violationsExample = templates.Examples(i18n.T(`
		# Get a list of policy violations resulting in request denial across all namespaces
		bbctl violations 
		
		# Get a list of policy violations resulting in request denial in the given namespace.
		bbctl violations -n NAMESPACE		
		
		# Get a list of policy violations reported by audit process across all namespaces
		bbctl violations --audit	
		
		# Get a list of policy violations reported by audit process in the given namespace
		bbctl violations --audit --namespace NAMESPACE	
	`))
)

// The strings should be passed to the combra.Command instantiation as follows:
func NewViolationsCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     violationsUse,
		Short:   violationsShort,
		Long:    violationsLong,
		Example: violationsExample,
		RuEn: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(getViolations(cmd, factory, streams))
		},
	}
}
```

