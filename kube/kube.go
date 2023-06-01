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
func NewClientConfig(cfg *config.K8S, auth *config.AuthInfo) (*clientcmdapi.Config, error) {
	kubeCfg := clientcmdapi.NewConfig()
	clusterConfig := clientcmdapi.Cluster{
		Server: cfg.Server,
	}

	if cfg.SkipTLS {
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

	kubeCfg.Clusters[cfg.ClusterName] = &clusterConfig
	kubeCfg.AuthInfos[cfg.AuthInfoName] = &clientcmdapi.AuthInfo{
		Token: auth.Token,
	}
	ctx := &clientcmdapi.Context{
		Cluster:  cfg.ClusterName,
		AuthInfo: cfg.AuthInfoName,
	}
	if cfg.Namespace != "" {
		ctx.Namespace = cfg.Namespace
	}
	kubeCfg.Contexts[cfg.ContextName] = ctx
	kubeCfg.CurrentContext = cfg.ContextName

	return kubeCfg, nil
}

// NewRestConfig returns a new rest config.
func NewRestConfig(cfg *config.K8S, auth *config.AuthInfo) (*rest.Config, error) {
	kubeCfg, err := NewClientConfig(cfg, auth)
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
