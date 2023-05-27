package main

import (
	"bytes"
	"text/template"
)

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
