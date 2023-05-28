package kube

import (
	"encoding/base64"
	"testing"

	"github.com/appleboy/deploy-k8s/config"
)

func TestNewKubeClientConfig(t *testing.T) {
	cfg := &config.K8S{
		Server:       "https://my-kubernetes-api-server",
		SkipTLS:      false,
		CaCert:       base64.StdEncoding.EncodeToString([]byte("base64-encoded-ca-cert")),
		ProxyURL:     "http://proxy.example.com",
		Namespace:    "my-namespace",
		ClusterName:  "my-cluster",
		AuthInfoName: "my-auth-info",
		ContextName:  "my-context",
	}

	auth := &config.AuthInfo{
		Token: base64.StdEncoding.EncodeToString([]byte("base64-encoded-token")),
	}

	kubeCfg, err := NewClientConfig(cfg, auth)
	if err != nil {
		t.Errorf("Error creating Kubernetes client config: %s", err)
		return
	}

	cluster, ok := kubeCfg.Clusters[cfg.ClusterName]
	if !ok {
		t.Errorf("Cluster '" + cfg.ClusterName + "' not found in the config")
		return
	}
	if cluster.Server != cfg.Server {
		t.Errorf("Expected server: %s, got: %s", cfg.Server, cluster.Server)
	}
	if cluster.InsecureSkipTLSVerify != false {
		t.Errorf("Expected InsecureSkipTLSVerify: false, got: %v", cluster.InsecureSkipTLSVerify)
	}

	authInfo, ok := kubeCfg.AuthInfos[cfg.AuthInfoName]
	if !ok {
		t.Errorf("AuthInfo '" + cfg.AuthInfoName + "' not found in the config")
		return
	}
	if authInfo.Token != "base64-encoded-token" {
		t.Errorf("Expected token: base64-encoded-token, got: %s", authInfo.Token)
	}

	context, ok := kubeCfg.Contexts[cfg.ContextName]
	if !ok {
		t.Errorf("Context '" + cfg.ContextName + "' not found in the config")
		return
	}
	if context.Cluster != cfg.ClusterName {
		t.Errorf("Expected cluster: %s, got: %s", cfg.ClusterName, context.Cluster)
	}
	if context.AuthInfo != cfg.AuthInfoName {
		t.Errorf("Expected authInfo: %s, got: %s", cfg.AuthInfoName, context.AuthInfo)
	}
	if context.Namespace != cfg.Namespace {
		t.Errorf("Expected namespace: %s, got: %s", cfg.Namespace, context.Namespace)
	}

	if kubeCfg.CurrentContext != cfg.ContextName {
		t.Errorf("Expected currentContext: %s, got: %s", cfg.ContextName, kubeCfg.CurrentContext)
	}
}
