package main

import (
	"encoding/base64"
	"testing"
)

func TestNewKubeClientConfig(t *testing.T) {
	cfg := &Config{
		Server:    "https://my-kubernetes-api-server",
		SkipTLS:   false,
		CaCert:    base64.StdEncoding.EncodeToString([]byte("base64-encoded-ca-cert")),
		ProxyURL:  "http://proxy.example.com",
		Token:     base64.StdEncoding.EncodeToString([]byte("base64-encoded-token")),
		Namespace: "my-namespace",
	}

	kubeCfg, err := NewKubeClientConfig(cfg)
	if err != nil {
		t.Errorf("Error creating Kubernetes client config: %s", err)
		return
	}

	// 驗證 cluster 設定
	cluster, ok := kubeCfg.Clusters["default"]
	if !ok {
		t.Errorf("Cluster 'default' not found in the config")
		return
	}
	if cluster.Server != cfg.Server {
		t.Errorf("Expected server: %s, got: %s", cfg.Server, cluster.Server)
	}
	if cluster.InsecureSkipTLSVerify != false {
		t.Errorf("Expected InsecureSkipTLSVerify: false, got: %v", cluster.InsecureSkipTLSVerify)
	}
	// 驗證其他 cluster 設定...

	// 驗證 authInfo 設定
	authInfo, ok := kubeCfg.AuthInfos["default"]
	if !ok {
		t.Errorf("AuthInfo 'default' not found in the config")
		return
	}
	if authInfo.Token != "base64-encoded-token" {
		t.Errorf("Expected token: base64-encoded-token, got: %s", authInfo.Token)
	}
	// 驗證其他 authInfo 設定...

	// 驗證 context 設定
	context, ok := kubeCfg.Contexts["default"]
	if !ok {
		t.Errorf("Context 'default' not found in the config")
		return
	}
	if context.Cluster != "default" {
		t.Errorf("Expected cluster: default, got: %s", context.Cluster)
	}
	if context.AuthInfo != "default" {
		t.Errorf("Expected authInfo: default, got: %s", context.AuthInfo)
	}
	if context.Namespace != cfg.Namespace {
		t.Errorf("Expected namespace: %s, got: %s", cfg.Namespace, context.Namespace)
	}
	// 驗證其他 context 設定...

	if kubeCfg.CurrentContext != "default" {
		t.Errorf("Expected currentContext: default, got: %s", kubeCfg.CurrentContext)
	}
}
