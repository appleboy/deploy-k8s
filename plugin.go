package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

var (
	reDronePlugin  = regexp.MustCompile(`^PLUGIN_(.*)=(.*)`)
	reDroneVar     = regexp.MustCompile(`^(DRONE_.*)=(.*)`)
	reGitHubAction = regexp.MustCompile(`^INPUT_(.*)=(.*)`)
	reGitHubVar    = regexp.MustCompile(`^(GITHUB_.*)=(.*)`)
)

// GetAllEnviroment returns all environment variables.
func GetAllEnviroment() map[string]any {
	envs := make(map[string]any)
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
		Templates []string
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

	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	kubeObjs, err := ParseTemplateSet(p.Config.Templates, GetAllEnviroment())

	for _, v := range kubeObjs {
		mapping, err := mapper.RESTMapping(v.GVK.GroupKind(), v.GVK.Version)
		if err != nil {
			return err
		}

		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if v.Obj.GetNamespace() == "" {
				if p.Config.Namespace == "" {
					return fmt.Errorf(
						"apply resource failed: namespace must be defined, apiVersion=%s, kind=%s, name=%s",
						v.GVK.GroupVersion().String(), v.GVK.Kind, v.Obj.GetName(),
					)
				}
				// set default namespace
				v.Obj.SetNamespace(p.Config.Namespace)
			}
			// namespaced resources should specify the namespace
			dr = dyn.
				Resource(mapping.Resource).
				Namespace(v.Obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = dyn.Resource(mapping.Resource)
		}

		obj, err := dr.Apply(
			context.Background(),
			v.Obj.GetName(),
			v.Obj,
			metav1.ApplyOptions{},
		)
		if err != nil {
			return err
		}

		log.Printf("filePath: %#v", v.TplPath)
		log.Printf("apiVersion: %#v", v.GVK.GroupVersion().String())
		log.Printf("kind: %#v", v.GVK.Kind)
		log.Printf("namespace: %#v", obj.GetNamespace())
		log.Printf("name: %#v", obj.GetName())
	}

	return nil
}
