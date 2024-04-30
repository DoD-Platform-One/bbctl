package config

import (
	"github.com/spf13/cobra"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

// ClientGetter is an struct for getting a BB config client.
type ClientGetter struct{}

// GetClient returns a new config client.
func (clientGetter *ClientGetter) GetClient(command *cobra.Command, loggingClient *bbLog.Client) (*ConfigClient, error) {
	return NewClient(
		getConfig,
		SetAndBindFlag,
		loggingClient,
		command,
	)
}
