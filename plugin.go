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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
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
var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

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
		Template  string
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

	format, err := os.ReadFile(p.Config.Template)
	if err != nil {
		return err
	}

	tpl, err := NewTemplateByString(string(format), GetAllEnviroment())
	if err != nil {
		return err
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(tpl), nil, obj)
	if err != nil {
		return err
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if obj.GetNamespace() == "" {
			if p.Config.Namespace == "" {
				return fmt.Errorf(
					"apply resource failed: namespace must be defined, apiVersion=%s, kind=%s, name=%s",
					gvk.GroupVersion().String(), gvk.Kind, obj.GetName(),
				)
			}
			// set default namespace
			obj.SetNamespace(p.Config.Namespace)
		}
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	engine2, err := dr.Apply(context.Background(), obj.GetName(), obj, metav1.ApplyOptions{
		FieldManager: "sample-controller",
	})

	log.Printf("apiVersion: %#v", gvk.GroupVersion().String())
	log.Printf("kind: %#v", gvk.Kind)
	log.Printf("namespace: %#v", engine2.GetNamespace())
	log.Printf("name: %#v", engine2.GetName())
	log.Printf("%#v", engine2.GetLabels())

	return nil
}
