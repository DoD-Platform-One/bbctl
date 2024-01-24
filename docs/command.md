# Command Symantics

bbctl draws upon the command syntax from the practices followed in kubectl.

## Syntax

Use the following syntax to run bbctl commands from your terminal window:

```
bbctl [COMMAND] [TYPE] [NAME] [FLAGS]
```

where COMMAND, TYPE, NAME, and FLAGS are:

* __COMMAND__: Specifies the operation that you want to perform on one or more resources, for example list, get. This applies in situations where we have a specific resource and the usage of REST style verb and resource is appropriate. However, in some cases, the command does not apply to a specific resource but the BigBang deployment as a whole, so using the operation name is enough, e.g., querying the status of current Bigbang deployment:
    ```
        bbctl status
    ```
* __TYPE__: Specifies the resource type. Resource types are case-insensitive and you can specify the singular, plural, or abbreviated forms. Currently there are no CRDs defined for BigBang, as such TYPE is not used in commands implemented as of writing of this version of the document.
* __NAME__: Specifies the name of the resource when TYPE is used. When the TYPE is ommited, NAME uniquely identifies a concept in the context of the command, e.g., querying the values for a helm release deployed by BigBang using the following command implies the name of a helm release:
    ```
        bbctl values gatekeeper-system-gatekeeper
    ```

* __FLAGS__: Specifies optional flags. For example, you can use the --kubeconfig flag to explicitly specify the location of kube config file. Some flags are specific to a given specific bbctl commands while other flags like --namespace and --kubeconfig are available for all the bbctl commands, e.g., querying the gatekeeper violations in audit mode:
    ```
        bbctl violations --audit
    ```