package http

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"
)

func ParseTarget(target string) (string, string) {
	parts := strings.SplitN(target, " ", 2)
	return parts[0], parts[1]
}

func RenderTemplate(text string, data *map[string]string) string {
	tmpl, err := template.New("tmpl").Parse(text)
	if err != nil {
		return text
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return text
	}
	return buf.String()
}

func GetValueFromPath(data map[string]any, path string) (any, bool) {
	keys := strings.Split(path, ".")
	var current any = data
	for _, key := range keys {
		if key == "$" {
			continue
		}
		if m, ok := current.(map[string]any); ok {
			current, ok = m[key]
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}

func MakeRequest(method, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	client := &http.Client{}
	return client.Do(req)
}
