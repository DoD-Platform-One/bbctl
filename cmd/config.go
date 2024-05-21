package cmd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
)

var (
	configUse = `config [key]`

	configShort = i18n.T(`Print bbctl configuration information.`)

	configLong = templates.LongDesc(i18n.T(`
		Output the current bbctl configurations set in the bbctl configuration file.
		
		Configurations are printed in the format:
			key: value

		To print a specific configuration, pass it as a keyword paramater to the "bbctl config" invocation.
		Example:
			$ bbctl config log-level
			info
	`))
)

// NewConfigCmd - create a new Cobra config command
func NewConfigCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	var err error
	cmd := &cobra.Command{
		Use:                   configUse,
		Short:                 configShort,
		Long:                  configLong,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(func(factory bbUtil.Factory, args []string) error {
				output, err := getBBConfig(cmd, factory, args)
				fmt.Println(output)
				return err
			}(factory, args))
		},
	}
	factory.GetLoggingClient().HandleError("Unable to local configuration:", err)

	return cmd
}

// getValueAsString returns a string representation of the value, handling bool values properly
func getValueAsString(field reflect.Value) string {
	if field.Kind() == reflect.Bool {
		return fmt.Sprintf("%t", field.Bool())
	}
	return field.String()
}

func findConfig(config any, key string) (string, error) {
	keys := strings.Split(key, ".")
	return findRecursive(reflect.ValueOf(config), keys)
}

func findRecursive(v reflect.Value, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("invalid key")
	}

	fieldName := keys[0]
	typ := v.Type()

	// If the input is a pointer, dereference it
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		typ = typ.Elem()
	}

	// Iterate through fields to find match
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("yaml") == fieldName {
			// If it's the last key, return the value
			if len(keys) == 1 {
				return getValueAsString(v.Field(i)), nil
			}
			// If it's a struct, recursively search within it
			if v.Field(i).Kind() == reflect.Struct {
				return findRecursive(v.Field(i), keys[1:])
			}
		}
	}

	return "", fmt.Errorf("No such field: %s", fieldName)
}

func getBBConfig(cmd *cobra.Command, factory bbUtil.Factory, args []string) (string, error) {
	// Fetch the global config struct
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return "", fmt.Errorf("error getting config client: %w", err)
	}

	config := configClient.GetConfig()

	switch len(args) {

	case 0:
		// If all keys are requested, we can just dump the config YAML wholesale
		// as this is probably what the user wants anyway
		// TODO - it would be easy here to implement an `-o` flag for JSON, YAML, etc...
		configBytes, err := yaml.Marshal(config)
		if err != nil {
			// This can't be tested because the yaml.Marshal function is a pain to mock and getting it to return an error naturally is difficult
			return "", fmt.Errorf("error marshaling global config: %w", err)
		}
		return strings.TrimSpace(string(configBytes)), nil

	// For an individual key, we need to use reflection (🥴) to try and locate
	// the value in our struct
	case 1:
		key := args[0]
		return findConfig(config, key)

	default:
		return "", errors.New("too many arguments passed to bbctl config")
	}
}
