package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type elasticsearchConnector struct {
	client *elasticsearch.Client
	host   string
}

func (c *elasticsearchConnector) Connect(configJSON string) error {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return fmt.Errorf("invalid config json: %w", err)
	}
	host, _ := cfg["host"].(string)
	if host == "" {
		host = "http://localhost:9200"
	}
	username, _ := cfg["username"].(string)
	password, _ := cfg["password"].(string)

	esCfg := elasticsearch.Config{
		Addresses: []string{host},
	}
	if username != "" {
		esCfg.Username = username
		esCfg.Password = password
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return fmt.Errorf("es client creation failed: %w", err)
	}
	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("es ping failed: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es ping returned error: %s", res.String())
	}
	c.client = client
	c.host = host
	return nil
}

func (c *elasticsearchConnector) Close() error {
	return nil // ES client 不需要显式关闭
}

func (c *elasticsearchConnector) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	// query 格式: "index_name|{"query": {"match_all": {}}}"
	parts := strings.SplitN(query, "|", 2)
	index := parts[0]
	body := `{"query":{"match_all":{}}}`
	if len(parts) > 1 && parts[1] != "" {
		body = parts[1]
	}

	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(strings.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("es search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("es decode failed: %w", err)
	}

	rows := make([]map[string]interface{}, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		rows = append(rows, hit.Source)
	}
	return rows, nil
}

func (c *elasticsearchConnector) GetColumns(ctx context.Context, dbName, tableName string) ([]ColumnInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	res, err := c.client.Indices.GetMapping(
		c.client.Indices.GetMapping.WithIndex(tableName),
	)
	if err != nil {
		return nil, fmt.Errorf("es mapping failed: %w", err)
	}
	defer res.Body.Close()

	var mapping map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&mapping); err != nil {
		return nil, fmt.Errorf("es mapping decode failed: %w", err)
	}

	cols := make([]ColumnInfo, 0)
	// 解析 mapping 结构提取字段名
	if idxMap, ok := mapping[tableName].(map[string]interface{}); ok {
		if m, ok := idxMap["mappings"].(map[string]interface{}); ok {
			if props, ok := m["properties"].(map[string]interface{}); ok {
				for name, fieldDef := range props {
					colType := "text"
					if fd, ok := fieldDef.(map[string]interface{}); ok {
						if t, ok := fd["type"].(string); ok {
							colType = t
						}
					}
					cols = append(cols, ColumnInfo{Name: name, Type: colType})
				}
			}
		}
	}
	return cols, nil
}
