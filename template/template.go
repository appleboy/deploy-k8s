package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializeryaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

var serializer = serializeryaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

type KubeObject struct {
	TplData string
	TplPath string
	GVK     schema.GroupVersionKind
	Obj     *unstructured.Unstructured
}

func (k *KubeObject) PrettyString() string {
	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, []byte(k.TplData), "", "  ")
	return prettyJSON.String()
}

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

// NewTemplate returns a string by template.
func NewTemplate(format string, data map[string]interface{}) ([]byte, error) {
	t, err := template.New("message").Parse(format)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer

	if err := t.Execute(&tpl, map[string]any{
		"envs": data,
	}); err != nil {
		return nil, err
	}

	return tpl.Bytes(), nil
}

// ParseSet returns a list of unstructured objects.
func ParseSet(templates []string, envMap map[string]any) ([]*KubeObject, error) {
	objects := make([]*KubeObject, 0)
	fileSets := []string{}

	for _, template := range templates {
		files, err := filepath.Glob(template)
		if err != nil {
			continue
		}
		fileSets = append(fileSets, files...)
	}

	for _, template := range fileSets {
		format, err := os.ReadFile(template)
		if err != nil {
			continue
		}

		tpl, err := NewTemplate(string(format), envMap)
		if err != nil {
			return nil, err
		}

		objs, err := ParseObject(tpl)
		if err != nil {
			return nil, err
		}

		for _, obj := range objs {
			objCopy := obj.DeepCopy()
			data, err := objCopy.MarshalJSON()
			if err != nil {
				return nil, err
			}

			objects = append(objects, &KubeObject{
				TplData: string(data),
				TplPath: template,
				GVK:     objCopy.GroupVersionKind(),
				Obj:     objCopy,
			})
		}
	}

	return objects, nil
}

// ParseObject returns a list of unstructured objects.
func ParseObject(data []byte) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 256)
	for {
		var rawObj pkgruntime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode data to raw object failed: %v", err)
		}

		var obj unstructured.Unstructured
		if err := pkgruntime.DecodeInto(serializer, rawObj.Raw, &obj); err != nil {
			return nil, fmt.Errorf("decode raw object failed: %v", err)
		}
		result = append(result, obj)
	}
	return result, nil
}
