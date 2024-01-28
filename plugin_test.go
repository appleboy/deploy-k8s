package main

import (
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
