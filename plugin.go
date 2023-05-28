package main

import (
	"context"
	"fmt"
	"log"

	"github.com/appleboy/deploy-k8s/config"
	"github.com/appleboy/deploy-k8s/kube"
	"github.com/appleboy/deploy-k8s/template"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	// Plugin values.
	Plugin struct {
		Config *config.K8S
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

	// Generate kube config
	if p.Config.Output != "" {
		kubeCfg, err := kube.NewClientConfig(p.Config)
		if err != nil {
			return err
		}
		err = clientcmd.WriteToFile(*kubeCfg, p.Config.Output)
		if err != nil {
			return err
		}
		log.Println("Generated kube config file:", p.Config.Output)
		return nil
	}

	restConfig, err := kube.NewRestConfig(p.Config)
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

	kubeObjs, err := template.ParseSet(p.Config.Templates, template.GetAllEnviroment())

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
			metav1.ApplyOptions{
				FieldManager: "deploy-k8s-plugin",
			},
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
