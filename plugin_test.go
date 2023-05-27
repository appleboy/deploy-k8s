package main

import (
	"os"
	"testing"
)

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
		"var1": "value1",
		"var2": "value2",
		"var3": "value3",
		"var4": "value4",
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
