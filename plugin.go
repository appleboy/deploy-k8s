package main

import (
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
	}

	// Plugin values.
	Plugin struct {
		Config *Config
	}
)

func (p *Plugin) Exec() error {
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
