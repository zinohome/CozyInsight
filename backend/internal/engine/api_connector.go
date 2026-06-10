package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type apiConnector struct {
	baseURL    string
	method     string
	headers    map[string]string
	params     map[string]string
	timeoutSec int
	jsonPath   string // 可选：JSON 结果中提取数据的路径，如 "data.items"
}

func (c *apiConnector) Connect(configJSON string) error {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}
	c.baseURL, _ = cfg["url"].(string)
	if c.baseURL == "" {
		return fmt.Errorf("missing required field: url")
	}
	c.method, _ = cfg["method"].(string)
	if c.method == "" {
		c.method = "GET"
	}
	if h, ok := cfg["headers"].(map[string]interface{}); ok {
		c.headers = make(map[string]string)
		for k, v := range h {
			if s, ok := v.(string); ok {
				c.headers[k] = s
			}
		}
	}
	if p, ok := cfg["params"].(map[string]interface{}); ok {
		c.params = make(map[string]string)
		for k, v := range p {
			if s, ok := v.(string); ok {
				c.params[k] = s
			}
		}
	}
	c.jsonPath, _ = cfg["jsonPath"].(string)
	if t, ok := cfg["timeout"].(float64); ok && t > 0 {
		c.timeoutSec = int(t)
	} else {
		c.timeoutSec = 30
	}
	return nil
}

func (c *apiConnector) Close() error { return nil }

func (c *apiConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	client := &http.Client{Timeout: time.Duration(c.timeoutSec) * time.Second}
	req, err := http.NewRequestWithContext(ctx, c.method, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("api request creation failed: %w", err)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if c.params != nil {
		q := req.URL.Query()
		for k, v := range c.params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api returned %d: %s", resp.StatusCode, string(body))
	}

	var raw interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("api json decode failed: %w", err)
	}

	// 根据 jsonPath 提取数据
	data := raw
	if c.jsonPath != "" {
		parts := strings.Split(c.jsonPath, ".")
		current := raw
		for _, part := range parts {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("jsonPath %q not navigable at segment %q", c.jsonPath, part)
			}
			current, ok = m[part]
			if !ok {
				return nil, fmt.Errorf("jsonPath %q not found (segment %q missing)", c.jsonPath, part)
			}
		}
		data = current
	}

	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("api response is not a JSON array at jsonPath=%q", c.jsonPath)
	}

	results := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			results = append(results, m)
		}
	}
	return results, nil
}

func (c *apiConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	// 先查询一次样本数据，提取字段名
	rows, err := c.Query(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("api sample query failed: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data to infer columns")
	}
	cols := make([]ColumnInfo, 0)
	for name := range rows[0] {
		cols = append(cols, ColumnInfo{Name: name, Type: "text"})
	}
	return cols, nil
}
