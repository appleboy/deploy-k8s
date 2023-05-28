package kube

import (
	"encoding/base64"
	"fmt"

	"github.com/appleboy/deploy-k8s/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// NewKubeClientConfig returns a new Kubernetes client config.
func NewClientConfig(cfg *config.K8S) (*clientcmdapi.Config, error) {
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

	return kubeCfg, nil
}

// NewRestConfig returns a new rest config.
func NewRestConfig(cfg *config.K8S) (*rest.Config, error) {
	kubeCfg, err := NewClientConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("new kube client config; %w", err)
	}
	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*kubeCfg, "default", &clientcmd.ConfigOverrides{}, nil)
	actualCfg, err := clientBuilder.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("client builder client config; %w", err)
	}

	return actualCfg, nil
}
