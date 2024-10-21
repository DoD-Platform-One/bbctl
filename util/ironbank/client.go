package ironbank

import (
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/credentialhelper"
)

const DefaultRegistryURI = "registry1.dso.mil"

type Client interface {
	GetImageSHA(image string) (string, error)
}

func NewClient(credentialHelper credentialhelper.CredentialHelper, getImageSHAFunc GetImageSHAFunc) (Client, error) {
	return &ironbankClient{
		credentialHelper: credentialHelper,
		getImageSHA:      getImageSHAFunc,
	}, nil
}

type ironbankClient struct {
	credentialHelper credentialhelper.CredentialHelper
	getImageSHA      GetImageSHAFunc
}

type GetImageSHAFunc func(credentialhelper.Credentials, string) (string, error)

// GetImageSHA looks up the remote image name provided utilizing the credentials provided
// and returns the trimmed sha256 digest of the image
func (c *ironbankClient) GetImageSHA(image string) (string, error) {
	username, err := c.credentialHelper("username", DefaultRegistryURI)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}
	password, err := c.credentialHelper("password", DefaultRegistryURI)
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	credentials := credentialhelper.Credentials{
		Username: username,
		Password: password,
	}

	return c.getImageSHA(credentials, image)
}
