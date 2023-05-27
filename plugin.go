package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var (
	reDronePlugin  = regexp.MustCompile(`^PLUGIN_(.*)=(.*)`)
	reDroneVar     = regexp.MustCompile(`^(DRONE_.*)=(.*)`)
	reGitHubAction = regexp.MustCompile(`^INPUT_(.*)=(.*)`)
	reGitHubVar    = regexp.MustCompile(`^(GITHUB_.*)=(.*)`)
)

// GetAllEnviroment returns all environment variables.
func GetAllEnviroment() map[string]string {
	envs := make(map[string]string)
	for _, e := range os.Environ() {
		// Drone CI
		if reDronePlugin.MatchString(e) {
			matches := reDronePlugin.FindStringSubmatch(e)
			key := strings.ToLower(matches[1])
			envs[key] = matches[2]
			continue
		}
		// Drone CI
		if reDroneVar.MatchString(e) {
			matches := reDroneVar.FindStringSubmatch(e)
			key := strings.ToLower(matches[1])
			envs[key] = matches[2]
			continue
		}
		// GitHub Actions
		if reGitHubAction.MatchString(e) {
			matches := reGitHubAction.FindStringSubmatch(e)
			key := strings.ToLower(matches[1])
			envs[key] = matches[2]
			continue
		}
		// GitHub Actions
		if reGitHubVar.MatchString(e) {
			matches := reGitHubVar.FindStringSubmatch(e)
			key := strings.ToLower(matches[1])
			envs[key] = matches[2]
			continue
		}
	}
	return envs
}

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
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.
		CoreV1().
		Pods(p.Config.Namespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	namespace := "All"
	if p.Config.Namespace != "" {
		namespace = p.Config.Namespace
	}
	fmt.Printf("[%s] There are %d pods in the cluster\n", namespace, len(pods.Items))

	_, err = dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	return nil
}
