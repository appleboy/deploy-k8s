package main

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type (
	// Config for the kube server.
	Config struct {
		Server    string
		SkipTLS   bool
		CaCert    string
		Token     string
		Namespace string
		ProxyURL  string
	}

	// Plugin values.
	Plugin struct {
		Config *Config
	}
)

func (p *Plugin) Exec() error {
	if p.Config.Server == "" {
		return fmt.Errorf("server is required")
	}
	if p.Config.Token == "" {
		return fmt.Errorf("token is required")
	}
	if p.Config.CaCert == "" {
		return fmt.Errorf("ca_cert is required")
	}

	restConfig, err := NewKubeConfig(p.Config)
	if err != nil {
		return err
	}
	_, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	_, err = dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	return nil
}
