package main

import (
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// NewKubeClient returns a new Kubernetes client.
func NewKubeConfig(cfg *Config) (*rest.Config, error) {
	kubeCfg := clientcmdapi.NewConfig()
	clusterConfig := clientcmdapi.Cluster{
		Server: cfg.Server,
	}

	if cfg.SkipTLS == true {
		clusterConfig.InsecureSkipTLSVerify = true
	} else {
		ca, err := base64.StdEncoding.DecodeString(cfg.CaCert)
		if err != nil {
			return nil, fmt.Errorf("possible corrupted CA, or not base64 encoded: %s", err)
		}
		clusterConfig.CertificateAuthorityData = ca
	}

	// Add proxy support
	if cfg.ProxyURL != "" {
		clusterConfig.ProxyURL = cfg.ProxyURL
	}

	token, err := base64.StdEncoding.DecodeString(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("invaild token, or not base64 encoded: %s", err)
	}

	kubeCfg.Clusters["default"] = &clusterConfig
	kubeCfg.AuthInfos["default"] = &clientcmdapi.AuthInfo{
		Token: string(token),
	}
	ctx := &clientcmdapi.Context{
		Cluster:  "default",
		AuthInfo: "default",
	}
	if cfg.Namespace != "" {
		ctx.Namespace = cfg.Namespace
	}
	kubeCfg.Contexts["default"] = ctx
	kubeCfg.CurrentContext = "default"
	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*kubeCfg, "default", &clientcmd.ConfigOverrides{}, nil)
	actualCfg, err := clientBuilder.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("client builder client config; %w", err)
	}

	return actualCfg, nil
}
