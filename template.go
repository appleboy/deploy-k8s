package main

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

type KubeObject struct {
	Tpl     string
	TplPath string
	GVK     *schema.GroupVersionKind
	Obj     *unstructured.Unstructured
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

// ParseTemplateSet returns a list of unstructured objects.
func ParseTemplateSet(templates []string, envMap map[string]any) ([]*KubeObject, error) {
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
		obj := &unstructured.Unstructured{}
		_, gvk, err := decUnstructured.Decode(tpl, nil, obj)
		if err != nil {
			return nil, err
		}

		objects = append(objects, &KubeObject{
			Tpl:     string(format),
			TplPath: template,
			GVK:     gvk,
			Obj:     obj,
		})
	}

	return objects, nil
}
