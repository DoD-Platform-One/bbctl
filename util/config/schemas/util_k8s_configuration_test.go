package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_UtilK8sConfiguration(t *testing.T) {
	var tests = []struct {
		desc string
		arg  *UtilK8sConfiguration
	}{
		{
			"reconcile configuration, no values",
			&UtilK8sConfiguration{},
		},
		{
			"reconcile configuration, kubeconfig set",
			&UtilK8sConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			cacheDir := "test1"
			kubeconfigPath := "test2"
			clusterName := "test3"
			authInfoName := "test4"
			context := "test5"
			namespace := "test6"
			apiServer := "test7"
			tlsServerName := "test8"
			insecure := true
			certFile := "test9"
			keyFile := "test10"
			caFile := "test11"
			bearerToken := "test12"
			impersonate := "test13"
			impersonateUID := "test14"
			impersonateGroup := []string{"test15"}
			username := "test16"
			password := "test17"
			timeout := "test18"
			disableCompression := true
			labelSelector := "test19"
			fieldSelector := "test20"
			allNamespaces := true
			all := true
			local := true
			instance := viper.New()
			instance.Set("kubeconfig", kubeconfigPath)
			instance.Set("cache-dir", cacheDir)
			instance.Set("cluster-name", clusterName)
			instance.Set("auth-info-name", authInfoName)
			instance.Set("context", context)
			instance.Set("namespace", namespace)
			instance.Set("api-server", apiServer)
			instance.Set("tls-server-name", tlsServerName)
			instance.Set("insecure", insecure)
			instance.Set("cert-file", certFile)
			instance.Set("key-file", keyFile)
			instance.Set("ca-file", caFile)
			instance.Set("bearer-token", bearerToken)
			instance.Set("impersonate", impersonate)
			instance.Set("impersonate-uid", impersonateUID)
			instance.Set("impersonate-group", impersonateGroup)
			instance.Set("username", username)
			instance.Set("password", password)
			instance.Set("timeout", timeout)
			instance.Set("disable-compression", disableCompression)
			instance.Set("label-selector", labelSelector)
			instance.Set("field-selector", fieldSelector)
			instance.Set("all-namespaces", allNamespaces)
			instance.Set("all", all)
			instance.Set("local", local)
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			assert.Equal(t, kubeconfigPath, tt.arg.Kubeconfig)
			assert.Equal(t, cacheDir, tt.arg.CacheDir)
			assert.Equal(t, clusterName, tt.arg.ClusterName)
			assert.Equal(t, authInfoName, tt.arg.AuthInfoName)
			assert.Equal(t, context, tt.arg.Context)
			assert.Equal(t, namespace, tt.arg.Namespace)
			assert.Equal(t, apiServer, tt.arg.APIServer)
			assert.Equal(t, tlsServerName, tt.arg.TLSServerName)
			assert.Equal(t, insecure, tt.arg.Insecure)
			assert.Equal(t, certFile, tt.arg.CertFile)
			assert.Equal(t, keyFile, tt.arg.KeyFile)
			assert.Equal(t, caFile, tt.arg.CAFile)
			assert.Equal(t, bearerToken, tt.arg.BearerToken)
			assert.Equal(t, impersonate, tt.arg.Impersonate)
			assert.Equal(t, impersonateUID, tt.arg.ImpersonateUID)
			assert.Equal(t, impersonateGroup, tt.arg.ImpersonateGroup)
			assert.Equal(t, username, tt.arg.Username)
			assert.Equal(t, password, tt.arg.Password)
			assert.Equal(t, timeout, tt.arg.Timeout)
			assert.Equal(t, disableCompression, tt.arg.DisableCompression)
			assert.Equal(t, labelSelector, tt.arg.LabelSelector)
			assert.Equal(t, fieldSelector, tt.arg.FieldSelector)
			assert.Equal(t, allNamespaces, tt.arg.AllNamespaces)
			assert.Equal(t, all, tt.arg.All)
			assert.Equal(t, local, tt.arg.Local)
		})
	}
}

func TestGetSubConfigurations_UtilK8sConfiguration(t *testing.T) {
	// Arrange
	arg := &UtilK8sConfiguration{}
	// Act
	result := arg.getSubConfigurations()
	// Assert
	assert.Equal(t, []BaseConfiguration{}, result)
}
