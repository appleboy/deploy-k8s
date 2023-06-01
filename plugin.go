package main

import (
	"context"
	"fmt"

	"github.com/appleboy/deploy-k8s/config"
	"github.com/appleboy/deploy-k8s/kube"
	"github.com/appleboy/deploy-k8s/template"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

type (
	// Plugin values.
	Plugin struct {
		Config   *config.K8S
		AuthInfo *config.AuthInfo
	}
)

func (p *Plugin) Exec() error {
	if p.Config.Server == "" {
		return fmt.Errorf("server is required")
	}
	if p.AuthInfo.Token == "" {
		return fmt.Errorf("token is required")
	}
	if p.Config.CaCert == "" {
		return fmt.Errorf("ca_cert is required")
	}

	// Generate kube config
	if p.Config.Output != "" {
		kubeCfg, err := kube.NewClientConfig(p.Config, p.AuthInfo)
		if err != nil {
			return err
		}
		err = clientcmd.WriteToFile(*kubeCfg, p.Config.Output)
		if err != nil {
			return err
		}
		log.Info().Str("file", p.Config.Output).Msg("Generated kube config file")
		return nil
	}

	restConfig, err := kube.NewRestConfig(p.Config, p.AuthInfo)
	if err != nil {
		return err
	}
	_, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	kubeObjs, err := template.ParseSet(p.Config.Templates, template.GetAllEnviroment())
	if err != nil {
		return err
	}

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
				Force:        true,
			},
		)
		if err != nil {
			return err
		}

		l := log.With().
			Str("apiVersion", v.GVK.GroupVersion().String()).
			Str("kind", v.GVK.Kind).
			Str("namespace", obj.GetNamespace()).
			Str("name", obj.GetName()).
			Logger()

		if p.Config.Debug {
			l.Debug().
				Str("template", v.TplPath).
				Msg("show resource")
			fmt.Printf("%s", v.PrettyString())
		}

		l.Info().
			Msg("apply resource success")
	}

	// update deployment container image
	deploymentRes := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	l := log.With().
		Str("namespace", p.Config.Namespace).
		Logger()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, err := dyn.Resource(deploymentRes).
			Namespace(p.Config.Namespace).
			Get(context.Background(), p.Config.Deployment, metav1.GetOptions{})
		if err != nil {
			return err
		}
		containers, found, err := unstructured.NestedSlice(result.Object, "spec", "template", "spec", "containers")
		if err != nil || !found || containers == nil {
			return fmt.Errorf("deployment containers not found or error in spec: %v", err)
		}
		for index, container := range containers {
			maps := container.(map[string]interface{})
			l.Info().Msgf("container name: %s, image name: %s", maps["name"], maps["image"])
			if maps["name"] == p.Config.Container {
				if err := unstructured.SetNestedField(
					containers[index].(map[string]interface{}),
					p.Config.Image,
					"image",
				); err != nil {
					return err
				}
			}
		}

		if err := unstructured.SetNestedField(
			result.Object, containers,
			"spec", "template", "spec", "containers",
		); err != nil {
			return err
		}

		_, err = dyn.Resource(deploymentRes).
			Namespace(p.Config.Namespace).
			Update(context.TODO(), result, metav1.UpdateOptions{
				FieldManager: "deploy-k8s-plugin",
			})
		if err != nil {
			return err
		}
		return nil
	})

	if retryErr != nil {
		return err
	}

	// list, err := dyn.Resource(deploymentRes).
	// 	Namespace(p.Config.Namespace).
	// 	List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	return err
	// }

	// for _, d := range list.Items {
	// 	replicas, found, err := unstructured.NestedInt64(d.Object, "spec", "replicas")
	// 	if err != nil || !found {
	// 		log.Warn().Err(err).Msgf(
	// 			"Replicas not found for deployment %s",
	// 			d.GetName())
	// 		continue
	// 	}
	// 	l.Info().
	// 		Int64("replicas", replicas).
	// 		Str("name", d.GetName()).
	// 		Msg("show replica number")
	// }

	return nil
}
