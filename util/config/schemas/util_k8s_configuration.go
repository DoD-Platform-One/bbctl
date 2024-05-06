package schemas

import "github.com/spf13/viper"

type UtilK8sConfiguration struct {
	// Cache directory: where to store the cache
	CacheDir string `mapstructure:"cache-dir" yaml:"cache-dir"`
	// kubeconfig file path: where to find the kubeconfig file
	Kubeconfig string `mapstructure:"kubeconfig" yaml:"kubeconfig"`

	// Config Flags

	// Cluster name: the name of the cluster
	ClusterName string `mapstructure:"cluster-name" yaml:"cluster-name"`
	// Auth info name: the name of the auth info
	AuthInfoName string `mapstructure:"auth-info-name" yaml:"auth-info-name"`
	// Context: the context
	Context string `mapstructure:"context" yaml:"context"`
	// Namespace: the namespace
	Namespace string `mapstructure:"namespace" yaml:"namespace"`
	// API server: the API server
	APIServer string `mapstructure:"api-server" yaml:"api-server"`
	// TLS server name: the TLS server name
	TLSServerName string `mapstructure:"tls-server-name" yaml:"tls-server-name"`
	// Insecure: whether to use an insecure connection
	Insecure bool `mapstructure:"insecure" yaml:"insecure"`
	// Cert file: the certificate file
	CertFile string `mapstructure:"cert-file" yaml:"cert-file"`
	// Key file: the key file
	KeyFile string `mapstructure:"key-file" yaml:"key-file"`
	// CA file: the CA file
	CAFile string `mapstructure:"ca-file" yaml:"ca-file"`
	// Bearer token: the bearer token
	BearerToken string `mapstructure:"bearer-token" yaml:"bearer-token"`
	// Impersonate: the user to impersonate
	Impersonate string `mapstructure:"impersonate" yaml:"impersonate"`
	// Impersonate UID: the UID to impersonate
	ImpersonateUID string `mapstructure:"impersonate-uid" yaml:"impersonate-uid"`
	// Impersonate group: the group to impersonate
	ImpersonateGroup []string `mapstructure:"impersonate-group" yaml:"impersonate-group"`
	// Username: the username
	Username string `mapstructure:"username" yaml:"username"`
	// Password: the password
	Password string `mapstructure:"password" yaml:"password"`
	// Timeout: the timeout
	Timeout string `mapstructure:"timeout" yaml:"timeout"`
	// Disable compression: whether to disable compression
	DisableCompression bool `mapstructure:"disable-compression" yaml:"disable-compression"`

	// Resource Builder Flags

	// Label selector: the label selector
	LabelSelector string `mapstructure:"label-selector" yaml:"label-selector"`
	// Field selector: the field selector
	FieldSelector string `mapstructure:"field-selector" yaml:"field-selector"`
	// All namespaces: whether to list the requested object(s) across all namespaces
	AllNamespaces bool `mapstructure:"all-namespaces" yaml:"all-namespaces"`
	// All: whether to list all
	All bool `mapstructure:"all" yaml:"all"`
	// Local: whether to list local
	Local bool `mapstructure:"local" yaml:"local"`
}

// ReconcileConfiguration reconciles the configuration.
func (u *UtilK8sConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("cache-dir") {
		u.CacheDir = instance.GetString("cache-dir")
	}
	if instance.IsSet("kubeconfig") {
		u.Kubeconfig = instance.GetString("kubeconfig")
	}

	// Config Flags
	if instance.IsSet("cluster-name") {
		u.ClusterName = instance.GetString("cluster-name")
	}
	if instance.IsSet("auth-info-name") {
		u.AuthInfoName = instance.GetString("auth-info-name")
	}
	if instance.IsSet("context") {
		u.Context = instance.GetString("context")
	}
	if instance.IsSet("namespace") {
		u.Namespace = instance.GetString("namespace")
	}
	if instance.IsSet("api-server") {
		u.APIServer = instance.GetString("api-server")
	}
	if instance.IsSet("tls-server-name") {
		u.TLSServerName = instance.GetString("tls-server-name")
	}
	if instance.IsSet("insecure") {
		u.Insecure = instance.GetBool("insecure")
	}
	if instance.IsSet("cert-file") {
		u.CertFile = instance.GetString("cert-file")
	}
	if instance.IsSet("key-file") {
		u.KeyFile = instance.GetString("key-file")
	}
	if instance.IsSet("ca-file") {
		u.CAFile = instance.GetString("ca-file")
	}
	if instance.IsSet("bearer-token") {
		u.BearerToken = instance.GetString("bearer-token")
	}
	if instance.IsSet("impersonate") {
		u.Impersonate = instance.GetString("impersonate")
	}
	if instance.IsSet("impersonate-uid") {
		u.ImpersonateUID = instance.GetString("impersonate-uid")
	}
	if instance.IsSet("impersonate-group") {
		u.ImpersonateGroup = instance.GetStringSlice("impersonate-group")
	}
	if instance.IsSet("username") {
		u.Username = instance.GetString("username")
	}
	if instance.IsSet("password") {
		u.Password = instance.GetString("password")
	}
	if instance.IsSet("timeout") {
		u.Timeout = instance.GetString("timeout")
	}
	if instance.IsSet("disable-compression") {
		u.DisableCompression = instance.GetBool("disable-compression")
	}

	// Resource Builder Flags
	if instance.IsSet("label-selector") {
		u.LabelSelector = instance.GetString("label-selector")
	}
	if instance.IsSet("field-selector") {
		u.FieldSelector = instance.GetString("field-selector")
	}
	if instance.IsSet("all-namespaces") {
		u.AllNamespaces = instance.GetBool("all-namespaces")
	}
	if instance.IsSet("all") {
		u.All = instance.GetBool("all")
	}
	if instance.IsSet("local") {
		u.Local = instance.GetBool("local")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (u *UtilK8sConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
