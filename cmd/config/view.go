package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

// NewConfigViewCmd - create a new Cobra config view command
func NewConfigViewCmd(factory bbUtil.Factory) *cobra.Command {
	var (
		viewUse   = `view [key]`
		viewShort = i18n.T(`Print bbctl configuration information.`)
		viewLong  = templates.LongDesc(i18n.T(`
			Output the current bbctl configurations set in the bbctl configuration file.
			
			Configurations are printed in the format:
				key: value
	
			To print a specific configuration, pass it as a keyword paramater to the "bbctl config" invocation.
			Example:
				$ bbctl config log-level
				info
		`))
	)

	cmd := &cobra.Command{
		Use:                   viewUse,
		Short:                 viewShort,
		Long:                  viewLong,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getBBConfig(cmd, factory, args)
		},
	}
	return cmd
}

// getValueAsString returns a string representation of the value, handling bool values properly
func getValueAsString(field reflect.Value) string {
	if field.Kind() == reflect.Bool {
		return strconv.FormatBool(field.Bool())
	}
	return field.String()
}

func findConfig(config any, key string) (string, error) {
	keys := strings.Split(key, ".")
	return findRecursive(reflect.ValueOf(config), keys)
}

func findRecursive(v reflect.Value, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", errors.New("invalid key")
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

	return "", fmt.Errorf("no such field: %s", fieldName)
}

func getBBConfig(cmd *cobra.Command, factory bbUtil.Factory, args []string) error {
	// Fetch the global config struct
	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return fmt.Errorf("error getting config client: %w", err)
	}

	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}

	outputClient, outClientErr := factory.GetOutputClient(cmd)
	if outClientErr != nil {
		return fmt.Errorf("error getting output client: %w", outClientErr)
	}

	switch len(args) {
	case 0:
		// If all keys are requested, we can just dump the config YAML wholesale
		// as this is probably what the user wants anyway
		outputErr := outputClient.Output(config)
		if outputErr != nil {
			return fmt.Errorf("error marshaling global config: %w", outputErr)
		}
		return nil

	// For an individual key, we need to use reflection (🥴) to try and locate
	// the value in our struct
	case 1:
		key := args[0]
		value, singleConfigErr := findConfig(config, key)
		if singleConfigErr != nil {
			return fmt.Errorf("error marshaling specific config: %w", singleConfigErr)
		}
		configValMap := map[string]interface{}{key: value}
		outputErr := outputClient.Output(
			&output.BasicOutput{
				Vals: configValMap,
			})
		if outputErr != nil {
			return fmt.Errorf("error creating output for specific config: %w", outputErr)
		}
		return nil

	default:
		return errors.New("too many arguments passed to bbctl config")
	}
}
