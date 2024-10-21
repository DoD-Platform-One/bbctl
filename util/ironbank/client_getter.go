package ironbank

import (
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/credentialhelper"
)

type ClientGetter struct{}

func (clientGetter *ClientGetter) GetClient(credentialHelper credentialhelper.CredentialHelper) (Client, error) {
	return NewClient(credentialHelper, getImageSHA)
}
