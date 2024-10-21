package ironbank

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/credentialhelper"
)

func getImageSHA(registryCredentials credentialhelper.Credentials, image string) (string, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return "", fmt.Errorf("failed to parse image reference: %w", err)
	}

	desc, err := remote.Get(ref, remote.WithAuth(authn.FromConfig(authn.AuthConfig{
		Username: registryCredentials.Username,
		Password: registryCredentials.Password,
	})))
	if err != nil {
		return "", fmt.Errorf("failed to get image description: %w", err)
	}

	sha := strings.TrimPrefix(desc.Digest.String(), "sha256:")
	return sha, nil
}
