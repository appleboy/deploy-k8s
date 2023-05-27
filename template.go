package main

import (
	"bytes"
	"text/template"
)

func NewTemplateByString(format string, data map[string]interface{}) (string, error) {
	t, err := template.New("message").Parse(format)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
