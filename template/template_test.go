package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	// Define test input
	format := "Hello, {{ .envs.name }}! Today is {{ .envs.day }}."

	data := map[string]interface{}{
		"name": "John",
		"day":  "Monday",
	}

	// Call the function being tested
	result, err := NewTemplate(format, data)
	// Check for errors
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	// Assert the expected result
	expected := "Hello, John! Today is Monday."
	if string(result) != expected {
		t.Errorf("Expected result: %s, but got: %s", expected, result)
	}
}

// TestGetAllEnviroment tests the GetAllEnviroment function.
func TestGetAllEnviroment(t *testing.T) {
	// Set up test environment variables
	os.Setenv("PLUGIN_VAR1", "value1")
	os.Setenv("DRONE_VAR2", "value2")
	os.Setenv("INPUT_VAR3", "value3")
	os.Setenv("GITHUB_VAR4", "value4")

	// Call the function being tested
	result := GetAllEnviroment()

	// Assert the expected values
	expected := map[string]string{
		"var1":        "value1",
		"drone_var2":  "value2",
		"var3":        "value3",
		"github_var4": "value4",
	}
	for key, expectedValue := range expected {
		actualValue, ok := result[key]
		if !ok {
			t.Errorf("Expected key %s not found in the result", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected value %s for key %s, but got %s", expectedValue, key, actualValue)
		}
	}
}

func TestParseObject(t *testing.T) {
	// 建立測試資料
	data := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap1
  namespace: testing
data:
  key1: value1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment1
  namespace: testing
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: myapp
        image: myimage
        env:
        - name: ENV_VAR
          value: test
`)

	objects, err := ParseObject(data)
	if err != nil {
		t.Errorf("Error parsing objects: %v", err)
		return
	}

	if len(objects) != 2 {
		t.Errorf("Expected 2 objects, got: %d", len(objects))
		return
	}

	obj1 := objects[0]
	if obj1.GetKind() != "ConfigMap" {
		t.Errorf("Expected Kind: ConfigMap, got: %s", obj1.GetKind())
	}
	if obj1.GetName() != "configmap1" {
		t.Errorf("Expected Name: configmap1, got: %s", obj1.GetName())
	}
	if obj1.GetNamespace() != "testing" {
		t.Errorf("Expected Name: testing, got: %s", obj1.GetName())
	}

	obj2 := objects[1]
	if obj2.GetKind() != "Deployment" {
		t.Errorf("Expected Kind: Deployment, got: %s", obj2.GetKind())
	}
	if obj2.GetName() != "deployment1" {
		t.Errorf("Expected Name: deployment1, got: %s", obj2.GetName())
	}
	if obj2.GetNamespace() != "testing" {
		t.Errorf("Expected Name: testing, got: %s", obj1.GetName())
	}
}

func TestParseSet(t *testing.T) {
	envMap := map[string]interface{}{
		"ENV_VAR": "test",
	}

	tempDir, err := os.MkdirTemp("", "test-templates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	template1 := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap1
data:
  key1: {{ .ENV_VAR }}
`)
	err = os.WriteFile(filepath.Join(tempDir, "template1.yaml"), template1, 0o644)
	if err != nil {
		t.Fatalf("Failed to write template1.yaml: %v", err)
	}

	template2 := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment1
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: myapp
        image: myimage
        env:
        - name: ENV_VAR
          value: {{ .ENV_VAR }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: drone-ci
`)
	err = os.WriteFile(filepath.Join(tempDir, "template2.yaml"), template2, 0o644)
	if err != nil {
		t.Fatalf("Failed to write template2.yaml: %v", err)
	}

	templates := []string{
		filepath.Join(tempDir, "template1.yaml"),
		filepath.Join(tempDir, "template2.yaml"),
	}
	objects, err := ParseSet(templates, envMap)
	if err != nil {
		t.Errorf("Error parsing objects: %v", err)
		return
	}

	if len(objects) != 3 {
		t.Errorf("Expected 2 objects, got: %d", len(objects))
		return
	}

	obj1 := objects[0]
	if obj1.TplPath != filepath.Join(tempDir, "template1.yaml") {
		t.Errorf("Expected TplPath: %s, got: %s", filepath.Join(tempDir, "template1.yaml"), obj1.TplPath)
	}
	if obj1.GVK.Group != "" || obj1.GVK.Version != "v1" || obj1.GVK.Kind != "ConfigMap" {
		t.Errorf("Expected GVK: v1/ConfigMap, got: %s/%s", obj1.GVK.GroupVersion(), obj1.GVK.Kind)
	}

	obj2 := objects[1]
	if obj2.TplPath != filepath.Join(tempDir, "template2.yaml") {
		t.Errorf("Expected TplPath: %s, got: %s", filepath.Join(tempDir, "template2.yaml"), obj2.TplPath)
	}
	if obj2.GVK.Group != "apps" || obj2.GVK.Version != "v1" || obj2.GVK.Kind != "Deployment" {
		t.Errorf("Expected GVK: apps/v1/Deployment, got: %s/%s", obj2.GVK.GroupVersion(), obj2.GVK.Kind)
	}

	obj3 := objects[2]
	if obj3.TplPath != filepath.Join(tempDir, "template2.yaml") {
		t.Errorf("Expected TplPath: %s, got: %s", filepath.Join(tempDir, "template2.yaml"), obj3.TplPath)
	}
	if obj3.GVK.Group != "" || obj3.GVK.Version != "v1" || obj3.GVK.Kind != "ServiceAccount" {
		t.Errorf("Expected GVK: v1/ServiceAccount, got: %s/%s", obj3.GVK.GroupVersion(), obj3.GVK.Kind)
	}
}
