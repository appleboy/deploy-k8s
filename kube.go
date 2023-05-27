package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/flowcontrol"
)

// NewKubeClient returns a new Kubernetes client.
func NewKubeConfig(cfg *Config) (*rest.Config, error) {
	kubeCfg := clientcmdapi.NewConfig()
	clusterConfig := clientcmdapi.Cluster{
		Server: cfg.Server,
	}

	if cfg.SkipTLS == true {
		clusterConfig.InsecureSkipTLSVerify = true
		log.Println("InsecureSkipTLSVerify flag set")
	}

	if cfg.CaCert != "" {
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

	kubeCfg.Clusters["default"] = &clusterConfig
	kubeCfg.AuthInfos["default"] = &clientcmdapi.AuthInfo{
		Token: cfg.Token,
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

	actualCfg.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(1000, 1000)
	return actualCfg, nil
}
