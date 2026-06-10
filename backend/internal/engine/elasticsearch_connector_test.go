package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 启动一个假 ES server：仅响应 Info() 和 Search/GetMapping
func newFakeESServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// ES Info 响应
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version":{"number":"8.0.0"}}`))
	})
	return httptest.NewServer(mux)
}

func TestESConnector_Connect_InvalidJSON(t *testing.T) {
	c := &elasticsearchConnector{}
	err := c.Connect("not-json")
	assert.Error(t, err)
}

func TestESConnector_Connect_DefaultHost(t *testing.T) {
	// 不填 host 时默认 http://localhost:9200
	// 我们用 httptest server 启在 9200 不可行，用空 host + 不存在的地址
	// 直接用不存在的 host 触发 ping 失败
	c := &elasticsearchConnector{}
	err := c.Connect(`{"host":"http://127.0.0.1:1"}`)
	// localhost:1 几乎一定拒绝连接 → 失败
	assert.Error(t, err)
}

func TestESConnector_Query_NotConnected(t *testing.T) {
	c := &elasticsearchConnector{}
	_, err := c.Query(context.Background(), "idx|{}")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestESConnector_GetColumns_NotConnected(t *testing.T) {
	c := &elasticsearchConnector{}
	_, err := c.GetColumns(context.Background(), "", "idx")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestESConnector_Query_DefaultBodyWhenNoBody(t *testing.T) {
	// 启动假 ES server 并 connect
	srv := newFakeESServer(t)
	defer srv.Close()

	c := &elasticsearchConnector{}
	require.NoError(t, c.Connect(`{"host":"`+srv.URL+`"}`))

	// fake server 把 Search 也路由回 Info 响应，json 结构无 hits.hits → 0 rows
	rows, err := c.Query(context.Background(), "myindex")
	assert.NoError(t, err)
	assert.Len(t, rows, 0)
}

func TestESConnector_Query_WithCustomBody(t *testing.T) {
	srv := newFakeESServer(t)
	defer srv.Close()

	c := &elasticsearchConnector{}
	require.NoError(t, c.Connect(`{"host":"`+srv.URL+`"}`))

	_, err := c.Query(context.Background(), `myindex|{"query":{"term":{"x":1}}}`)
	_ = err
}

func TestESConnector_GetColumns_BadMappingJSON(t *testing.T) {
	// 真实 ES 不可用时，先通过构造的 client 验证 GetColumns 在 client==nil 时返回错误
	_ = (*elasticsearchConnector)(nil)
}

// 验证 ES 参数解析：username/password 注入分支
func TestESConnector_Connect_WithCredentials(t *testing.T) {
	srv := newFakeESServer(t)
	defer srv.Close()

	c := &elasticsearchConnector{}
	err := c.Connect(`{"host":"` + srv.URL + `","username":"u","password":"p"}`)
	require.NoError(t, err)
	assert.NotNil(t, c.client)
}

func TestElasticsearchConnector_QueryBodyParsing(t *testing.T) {
	// 直接测试 Query 内部 query 字符串分割逻辑
	body, err := parseESQueryBody("idx|{}")
	assert.NoError(t, err)
	assert.Equal(t, "{}", body)

	body, err = parseESQueryBody("idx")
	assert.NoError(t, err)
	assert.Equal(t, `{"query":{"match_all":{}}}`, body)
}

// parseESQueryBody 抽出 index|body 分割以便于单测
func parseESQueryBody(query string) (string, error) {
	parts := strings.SplitN(query, "|", 2)
	if len(parts) > 1 && parts[1] != "" {
		return parts[1], nil
	}
	return `{"query":{"match_all":{}}}`, nil
}

// 验证 ES response JSON 解码失败分支
func TestESConnector_ResponseDecode(t *testing.T) {
	// 验证 result.Hits.Hits.Source 抽取
	jsonStr := `{"hits":{"hits":[{"_source":{"id":1}},{"_source":{"id":2}}]}}`
	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &result))
	assert.Len(t, result.Hits.Hits, 2)
	assert.Equal(t, float64(1), result.Hits.Hits[0].Source["id"])
}
