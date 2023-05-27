package main

import "testing"

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
