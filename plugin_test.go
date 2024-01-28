package main

import (
	"os"
	"testing"

	"github.com/appleboy/deploy-k8s/config"
)

func TestCheckConfig(t *testing.T) {
	// Create a new instance of the Plugin struct
	p := &Plugin{
		Config: &config.K8S{
			Server: "example.com",
			Output: "kubeconfig.yaml",
		},
		AuthInfo: &config.AuthInfo{
			Token: "abc123",
		},
	}

	// Test case: server is empty
	p.Config.Server = ""
	err := p.Exec()
	if err == nil || err.Error() != "server is required" {
		t.Errorf("Expected error: server is required")
	}

	// Test case: token is empty
	p.Config.Server = "example.com"
	p.AuthInfo.Token = ""
	err = p.Exec()
	if err == nil || err.Error() != "token is required" {
		t.Errorf("Expected error: token is required")
	}
}

func TestKubeConfig(t *testing.T) {
	// Create a new instance of the Plugin struct
	p := &Plugin{
		Config: &config.K8S{
			Server: os.Getenv("K8S_SERVER"),
			CaCert: os.Getenv("K8S_CA_CERT"),
			Output: "kubeconfig.yaml",
			Debug:  true,
		},
		AuthInfo: &config.AuthInfo{
			Token: os.Getenv("K8S_TOKEN"),
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Expected error: cannot create kube config: %s", err.Error())
	}
}

func TestDeployContainer(t *testing.T) {
	// Create a new instance of the Plugin struct
	p := &Plugin{
		Config: &config.K8S{
			Server:       os.Getenv("K8S_SERVER"),
			CaCert:       os.Getenv("K8S_CA_CERT"),
			Debug:        true,
			Namespace:    "test-namespace",
			ContextName:  "test-context",
			AuthInfoName: "test-authinfo",
			ClusterName:  "test-cluster",
			Templates:    []string{"testdata/deployment01.yaml"},
		},
		AuthInfo: &config.AuthInfo{
			Token: os.Getenv("K8S_TOKEN"),
		},
	}

	err := p.Exec()
	if err != nil {
		t.Errorf("Expected error: cannot deploy by template: %s", err.Error())
	}
}
