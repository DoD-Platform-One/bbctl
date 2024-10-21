package ironbank

import (
	"fmt"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/ironbank"
)

type GetImageSHAFunc func(image string) (string, error)

func NewFakeClient(getImageSHAFunc GetImageSHAFunc) ironbank.Client {
	return &FakeClient{
		getImageSHA: getImageSHAFunc,
	}
}

type FakeClient struct {
	getImageSHA GetImageSHAFunc
}

func (c *FakeClient) GetImageSHA(image string) (string, error) {
	if c.getImageSHA != nil {
		return c.getImageSHA(image)
	}
	return fmt.Sprintf("1234567890"), nil
}
