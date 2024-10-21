package credentialhelper

// CredentialsFile struct represents credentials YAML files with a top level field called `credentials`
// which contains a list of Credentials struct values
type CredentialsFile struct {
	Credentials []Credentials `yaml:"credentials"`
}

// Credentials struct represents each individual entry in a valid credentials file.
// URI should be unique for every entry
type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URI      string `yaml:"uri"`
}

// CredentialHelper is a function type that can be used to fetch credential values
// The function takes 2 parameters:
//
// * component (string) - The Credentials struct field name, either `username` or `password`
//
// * uri (string) - The Credentials struct URI value which uniquely identifies the requested component
//
// These parameters are passed into custom credential helpers as CLI arguments in the same order.
type CredentialHelper func(component string, uri string) (string, error)
