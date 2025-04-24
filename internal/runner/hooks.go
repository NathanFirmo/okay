package runner

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/NathanFirmo/okay/internal/http"
)

func (r *Runner) execHooks() error {
	for _, hook := range r.suite.BeforeEach.Hooks {
		method, url := http.ParseTarget(hook.Target)
		url = http.RenderTemplate(url, &r.captures)
		headers := make(map[string]string)

		for k, v := range hook.Headers {
			headers[k] = http.RenderTemplate(v, &r.captures)
		}

		resp, err := http.MakeRequest(method, url, headers)
		if err != nil {
			return fmt.Errorf("error making hook request: %v", err)
		}

		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var data map[string]any
		json.Unmarshal(body, &data)

		for key, path := range hook.Captures {
			value, ok := http.GetValueFromPath(data, path)
			if ok {
				r.captures[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return nil
}
