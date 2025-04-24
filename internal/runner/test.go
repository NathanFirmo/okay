package runner

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/NathanFirmo/okay/internal/db"
	"github.com/NathanFirmo/okay/internal/http"
	"github.com/NathanFirmo/okay/internal/parser"
)

func (r *Runner) execTests() error {
	passed := 0
	failed := 0

	for _, test := range r.suite.Tests {
		db.RunSeeds(r.suite, r.databases)
		r.execHooks()
		r.exec(test, &passed, &failed)
	}

	fmt.Printf("\nSummary: %d passed, %d failed\n", passed, failed)

	return nil
}

func (r *Runner) exec(t parser.Test, passed *int, failed *int) error {
	var failures []string
	method, url := http.ParseTarget(t.Target)
	url = http.RenderTemplate(url, &r.captures)
	headers := make(map[string]string)
	for k, v := range t.Headers {
		headers[k] = http.RenderTemplate(v, &r.captures)
	}
	resp, err := http.MakeRequest(method, url, headers)
	if err != nil {
		failures = append(failures, fmt.Sprintf("Error making request: %v", err))
	} else {
		if resp.StatusCode != t.Expect.Status {
			failures = append(failures, fmt.Sprintf("Expected status %d, got %d", t.Expect.Status, resp.StatusCode))
		}
		if len(t.Expect.Body) > 0 {
			body, _ := io.ReadAll(resp.Body)
			var data map[string]any
			err = json.Unmarshal(body, &data)
			if err != nil {
				failures = append(failures, fmt.Sprintf("Error unmarshaling response body: %v", err))
			} else {
				for key, expected := range t.Expect.Body {
					actual, ok := http.GetValueFromPath(data, key)
					if !ok {
						failures = append(failures, fmt.Sprintf("Expected body.%s not found", key))
					} else if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
						failures = append(failures, fmt.Sprintf("Expected body.%s to be %v, got %v", key, expected, actual))
					}
				}
			}
		}
	}

	for idx, data := range t.DBChecks {
		r.checkDb(idx, data, &failures)
	}

	if len(failures) == 0 {
		fmt.Printf("%s PASS %s %s\n", greenBg, reset, t.Name)
		*passed++
	} else {
		fmt.Printf("%s FAIL %s %s\n", redBg, reset, t.Name)
		for _, msg := range failures {
			fmt.Printf("    %s%s%s\n", dim, msg, reset)
		}
		*failed++
	}

	return nil
}

func (r *Runner) checkDb(idx int, dbCheck parser.DBCheck, failures *[]string) {
	dbConn, ok := r.databases[dbCheck.DB]
	if !ok {
		*failures = append(*failures, fmt.Sprintf("Database alias not found: %s", dbCheck.DB))
		return
	}

	// Query to fetch all rows
	rows, err := dbConn.Query(dbCheck.Query)
	if err != nil {
		*failures = append(*failures, fmt.Sprintf("Error executing query: %v", err))
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		*failures = append(*failures, fmt.Sprintf("Error getting columns: %v", err))
		return
	}

	// Store all rows
	var allRows [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			*failures = append(*failures, fmt.Sprintf("Error scanning row: %v", err))
			continue
		}

		// Copy values to avoid overwriting
		rowValues := make([]any, len(columns))
		copy(rowValues, values)
		allRows = append(allRows, rowValues)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		*failures = append(*failures, fmt.Sprintf("Error during row iteration: %v", err))
		return
	}

	// Check if any rows were returned
	if len(allRows) == 0 {
		*failures = append(*failures, fmt.Sprintf("No rows returned for query: %s", dbCheck.Query))
		return
	}

	// Check if Expect is empty
	if len(dbCheck.Expect) == 0 {
		*failures = append(*failures, fmt.Sprintf("No expected results provided for query: %s", dbCheck.Query))
		return
	}

	// Iterate over rows and expected maps, matching by index
	for i := 0; i < len(allRows) && i < len(dbCheck.Expect); i++ {
		values := allRows[i]             // Row at index i
		expectedMap := dbCheck.Expect[i] // Expected map at index i

		for j, col := range columns {
			expected, ok := expectedMap[col]
			if !ok {
				continue
			}

			actual := values[j]
			switch v := actual.(type) {
			case nil:
				if expected != nil {
					*failures = append(*failures, fmt.Sprintf(
						"In database check %d, at row %d, expected %s to be %v, got nil", idx, i, col, expected))
				}
			case []byte:
				actualStr := string(v)
				if expStr, ok := expected.(string); ok {
					if actualStr != expStr {
						*failures = append(*failures, fmt.Sprintf(
							"In database check %d, at row %d, expected %s to be %s, got %s", idx, i, col, expStr, actualStr))
					}
				} else {
					*failures = append(*failures, fmt.Sprintf(
						"In database check %d, at row %d, type mismatch for %s: expected %T, got %T", idx, i, col, expected, v))
				}
			case string:
				if expStr, ok := expected.(string); ok {
					if v != expStr {
						*failures = append(*failures, fmt.Sprintf(
							"In database check %d, at row %d, expected %s to be %s, got %s", idx, i, col, expStr, v))
					}
				} else {
					*failures = append(*failures, fmt.Sprintf(
						"In database check %d, at row %d, type mismatch for %s: expected %T, got %T", idx, i, col, expected, v))
				}
			case int64:
				if expInt, ok := expected.(int); ok {
					if int(v) != expInt {
						*failures = append(*failures, fmt.Sprintf(
							"In database check %d, at row %d, expected %s to be %d, got %d", idx, i, col, expInt, v))
					}
				} else if expInt64, ok := expected.(int64); ok {
					if v != expInt64 {
						*failures = append(*failures, fmt.Sprintf(
							"In database check %d, at row %d, expected %s to be %d, got %d", idx, i, col, expInt64, v))
					}
				} else {
					*failures = append(*failures, fmt.Sprintf(
						"In database check %d, at row %d, type mismatch for %s: expected %T, got %T", idx, i, col, expected, v))
				}
			default:
				actualStr := fmt.Sprintf("%v", v)
				expectedStr := fmt.Sprintf("%v", expected)
				if actualStr != expectedStr {
					*failures = append(*failures, fmt.Sprintf(
						"In database check %d, at row %d, expected %s to be %s, got %s (type %T)", idx, i, col, expectedStr, actualStr, v))
				}
			}
		}
	}
}
