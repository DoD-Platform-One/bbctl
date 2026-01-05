# Adding New Helm Chart Commands

Currently the `bbctl` helm chart is configured to run a `bbctl` command as a cronjob resource e.g. `bbctl version -U -A`.
This document defines the process used to add additional Helm Chart resources.


## Creating a New Chart Job


  This is primarily focused on creating new resources or jobs pertaining to `bbctl` from scratch.

  ```bash
  $ cd ~/path/to/bbctl/
  $ helm create chart
  ```

 Add and adjust your variables and set up necessary resources. Change and adjust values.yaml file as necessary.
 Please refer to the official documentation [here](https://helm.sh/docs/helm/helm_create/).
 
 Here are some additional helpful links: 

- [Template Best Practices](https://helm.sh/docs/chart_best_practices/templates/)

- [Writing Cronjob Spec](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#writing-a-cronjob-spec)


## Adding New Command To Existing Cronjob Chart


  This will primarily be focused on adding new commands to our existing cronjob definitions.
  A script has been created to automate this process [here](../scripts/add-cronjob.sh).


Use the following syntax to add `bbctl` cronjob commands from your terminal window:

```
$ ./bbctl/scripts/add-cronjob.sh  [TEMPLATE_COMMAND] [NEW_COMMAND]
```

where `TEMPLATE_COMMAND`, and `NEW_COMMAND` are parameters passed into the script as such:

* __TEMPLATE_COMMAND__: Specifies the command to template our new cronjob resource definition after. A new folder under `/bbctl/chart/templates/bigbangTemplate` has been added, and is what should be used for most use cases where a new cronjob command needs to be added.

* __NEW_COMMAND__: Specifies the command to create from the template for our new cronjob resource definition. A new folder under `/bbctl/chart/templates/bigbang<NEW_COMMAND>` will be added, and will be the keyword replacing all instances of the previous argument as necessary. 


### Additional Notes/Considerations


- The script works regardless of the capitalization or casing of the parameters passed in, e.g.:
    `./scripts/add-cronjob.sh template status` is equivalent to `./scripts/add-cronjob.sh Template Status`, and so on.

- Adding labels and baseLabels can be done in `bbctl/chart/values.yaml` manually.

- Be sure to verify all additions and changes made are correct in the `bbctl/chart/values.yaml` file and in `bbctl/chart/templates` folder.

- By default, the new cronjob resource definition will be `enabled: true` in `bbctl/chart/values.yaml`. Disable any specific cronjob deployment in your own overrides `bbctl.yaml` file.


### Verify


- Verify that `bbctl/chart/templates/bigbang<NEW_COMMAND>/_helpers.tpl` has all the necessary arguments defined afterwards.

- Verify the newly created cronjob works on your bigbang deployment with full logs populating from stdout with helm upgrade:

      $ helm upgrade --install bigbang-status ./chart -n bbctl


#### Example


If we wish to use `template` as our template definition for a new cronjob resource to create `bbctl status` aka `bigbangStatus` job we would call the script as such:

      $ ./scripts/add-cronjob.sh template status

