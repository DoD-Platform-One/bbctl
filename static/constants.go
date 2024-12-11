package static

import (
	"embed"

	"repo1.dso.mil/big-bang/apps/developer-tools/go-utils/yamler"
)

var (
	//go:embed resources
	resources embed.FS
	constants = Constants{} //nolint:unused,gochecknoglobals

	buildDate = "UNSET" //nolint:gochecknoglobals
)

type Readable interface {
	readConstants() error
}

type ReadFileFunc func(string) ([]byte, error)

func (r ReadFileFunc) ReadFile(s string) ([]byte, error) {
	return r(s)
}

type Constants struct {
	// readFileFunc - function to read file
	readFileFunc ReadFileFunc
	// BigBangHelmReleaseName - Helm Release Name for Big Bang Deployment
	BigBangHelmReleaseName string `yaml:"BigBangHelmReleaseName"`
	// BigBangNamespace - Namespace where Big Bang Helm Chart is deployed
	BigBangNamespace string `yaml:"BigBangNamespace"`
	// BigBangCliVersion - constance for sematic versioning
	BigBangCliVersion string `yaml:"BigBangCliVersion"`
	// BigBangBuildDate - Build date of compiled resources
	BigBangBuildDate string `yaml:"BigBangBuildDate"`
}

func (c *Constants) readConstants() error {
	yamlFile, err := c.readFileFunc.ReadFile("resources/constants.yaml")
	if err != nil {
		return err
	}
	c.BigBangBuildDate = buildDate
	err = yamler.Unmarshal(yamlFile, c)
	return err
}

// ConstantsClient is an interface that defines methods to interact with Constants
type ConstantsClient interface {
	GetConstants() (Constants, error)
}

// constantsClient is an implementation of the ConstantsClient interface
type constantsClient struct {
	readFileFunc ReadFileFunc
}

// NewConstantsClient creates a new ConstantsClient with the provided ReadFileFunc
func NewConstantsClient(readFileFunc ReadFileFunc) ConstantsClient {
	return &constantsClient{
		readFileFunc: readFileFunc,
	}
}

// GetConstants reads the constants from the YAML file and returns a Constants instance
func (c *constantsClient) GetConstants() (Constants, error) {
	constants := Constants{
		readFileFunc: c.readFileFunc,
	}
	err := constants.readConstants()
	return constants, err
}

// Default client using embedded resources
var DefaultClient = NewConstantsClient(resources.ReadFile) //nolint:gochecknoglobals

// GetDefaultConstants returns constants using the default client
func GetDefaultConstants() (Constants, error) {
	return DefaultClient.GetConstants()
}
