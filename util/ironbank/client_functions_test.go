package ironbank

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/credentialhelper"
)

func TestGetImageSHA(t *testing.T) {
	tests := []struct {
		name        string
		credentials credentialhelper.Credentials
		image       string
		serverResp  func(w http.ResponseWriter, r *http.Request)
		wantSHA     string
		wantErr     bool
	}{
		{
			name: "Successful SHA retrieval",
			credentials: credentialhelper.Credentials{
				Username: "testuser",
				Password: "testpass",
			},
			image: "localhost:5000/myimage:latest",
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				// Respond to the initial API version check
				if r.URL.Path == "/v2/" {
					w.WriteHeader(http.StatusOK)
					return
				}

				// Error on anything else
				if r.URL.Path != "/v2/myimage/manifests/latest" {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}

				// Write some garbage back, this will be internally SHAd
				w.Header().Set("Docker-Content-Digest", "sha256:1234567890abcdef")
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				err := encoder.Encode(map[string]interface{}{
					"schemaVersion": 2,
					"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
				})
				assert.NoError(t, err)
			},
			// This was calculated by the server
			wantSHA: "8c70a933efd9403d2412a4db4de8e47c2e1dccd680a0efcde1625ee94ab5d1c9",
			wantErr: false,
		},
		{
			name: "Server error",
			credentials: credentialhelper.Credentials{
				Username: "testuser",
				Password: "testpass",
			},
			image: "localhost:5000/myimage:latest",
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/" {
					w.WriteHeader(http.StatusOK)
					return
				}

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			wantSHA: "",
			wantErr: true,
		},
	}

	credentialHelper := func(component, _ string) (string, error) {
		switch component {
		case "username":
			return "testuser", nil
		case "password":
			return "testpass", nil
		default:
			return "", fmt.Errorf("unknown credential component: %s", component)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverResp))
			defer server.Close()

			// Override the registry in the image reference to point to our test server
			image := fmt.Sprintf("%s/%s", server.URL[7:], tt.image[strings.Index(tt.image, "/")+1:])

			clientGetter := ClientGetter{}
			client, _ := clientGetter.GetClient(credentialHelper)

			gotSHA, err := client.GetImageSHA(image)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantSHA, gotSHA)
		})
	}
}

func TestGetImageSHABadCredentials(t *testing.T) {
	tests := []struct {
		name             string
		credentialHelper credentialhelper.CredentialHelper
		expectedErr      string
	}{
		{
			name: "No Username",
			credentialHelper: func(_, _ string) (string, error) {
				return "", errors.New("no username")
			},
			expectedErr: "failed to get username: no username",
		},
		{
			name: "No Password",
			credentialHelper: func(component, _ string) (string, error) {
				if component == "password" {
					return "", errors.New("no password")
				}
				return "no error", nil
			},
			expectedErr: "failed to get password: no password",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientGetter := ClientGetter{}
			client, err := clientGetter.GetClient(test.credentialHelper)
			require.NoError(t, err)

			gotSHA, err := client.GetImageSHA("localhost:5000/myimage:latest")

			assert.Equal(t, test.expectedErr, err.Error())
			assert.Equal(t, "", gotSHA)
		})
	}
}
