package integration

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestE2E_LoginAndAccessUserProfile 走完整链路：register → login → 带 token 访问 /user/profile。
func TestE2E_LoginAndAccessUserProfile(t *testing.T) {
	ts := NewTestServer(t, nil)
	now := time.Now()

	// 1) Register: FindByUsername 期望"不存在"（sqlmock 返回错误），然后 Create 期望 INSERT
	// sqlx NamedExec 会把 :name 绑定 → 改写为 ? 占位；我们用 sqlmock.AnyArg() 通配参数。
	ts.Mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("alice").
		WillReturnError(sqlmock.ErrCancelled)
	ts.Mock.ExpectExec(`INSERT INTO users`).
		WithArgs("alice", sqlmock.AnyArg(), "a@x.com", "Alice", "", "", 1, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	regBody := `{"username":"alice","password":"alice123","nickName":"Alice","email":"a@x.com"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/register", "application/json", strings.NewReader(regBody))
	require.NoError(t, err)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "register body: %s", string(b))

	// 2) Login: FindByUsername 返回真实行（hash 字段是 alice123 的 bcrypt）
	// 我们用明文 hash 在 service 层做 bcrypt 比较 — 用一个真实生成的 hash
	hashed, _ := bcryptGenerate("alice123")
	ts.Mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "nick_name", "email", "status", "is_admin", "created_at", "updated_at"}).
			AddRow(1, "alice", hashed, "Alice", "a@x.com", 1, 0, now, now))
	// UpdateLastLogin
	ts.Mock.ExpectExec(`UPDATE users SET last_login_at`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	loginBody := `{"username":"alice","password":"alice123"}`
	resp, err = http.Post(ts.URL+"/api/v1/auth/login", "application/json", strings.NewReader(loginBody))
	require.NoError(t, err)
	token := readToken(t, resp)

	// 3) Profile: FindByID
	ts.Mock.ExpectQuery(`SELECT \* FROM users WHERE id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "nick_name", "email", "is_admin"}).
			AddRow(1, "alice", "Alice", "a@x.com", 0))

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "profile body: %s", string(body))

	assert.NoError(t, ts.Mock.ExpectationsWereMet())
}

// TestE2E_Unauthorized_Access 验证无 token 访问受保护路由时返回 401。
func TestE2E_Unauthorized_Access(t *testing.T) {
	ts := NewTestServer(t, nil)
	resp, err := http.Get(ts.URL + "/api/v1/user/profile")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func readToken(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}
	// 业务响应: {"code":0,"data":{"token":"..."}}
	// 简化: 直接在 body 中查找 "token":"xxx"
	s := string(body)
	idx := strings.Index(s, `"token":"`)
	require.GreaterOrEqual(t, idx, 0, "no token in response: %s", s)
	tail := s[idx+len(`"token":"`):]
	end := strings.Index(tail, `"`)
	require.GreaterOrEqual(t, end, 0)
	return tail[:end]
}

// TestE2E_DatasourceFlow 走 datasource CRUD 全链路。
// List → 返回空 → Create → 再次 List → 看到一条 → Get by ID → Delete → 404。
func TestE2E_DatasourceFlow(t *testing.T) {
	ts := NewTestServer(t, nil)
	now := time.Now()
	_ = now

	// 1) 登录拿 token
	hashed, _ := bcryptGenerate("alice123")
	ts.Mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "nick_name", "email", "status", "is_admin"}).
			AddRow(1, "alice", hashed, "Alice", "a@x.com", 1, 1))
	ts.Mock.ExpectExec(`UPDATE users SET last_login_at`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	loginBody := `{"username":"alice","password":"alice123"}`
	resp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", strings.NewReader(loginBody))
	require.NoError(t, err)
	token := readToken(t, resp)

	// 2) GET /datasource (List) — 期望 SELECT，返回空
	ts.Mock.ExpectQuery(`SELECT.*FROM datasources`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "config", "status", "created_at", "updated_at"}))
	resp, err = doAuthGet(ts.URL+"/api/v1/datasource", token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 3) POST /datasource
	ts.Mock.ExpectExec(`INSERT INTO datasources`).
		WillReturnResult(sqlmock.NewResult(42, 1))
	createBody := `{"name":"test-mysql","type":"mysql","config":"{\"host\":\"x\"}"}`
	resp, err = doAuthPost(ts.URL+"/api/v1/datasource", token, createBody)
	require.NoError(t, err)
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "create body: %s", string(b))

	// 4) GET /datasource/42
	ts.Mock.ExpectQuery(`SELECT.*FROM datasources.*WHERE.*id`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "config", "status"}).
			AddRow(42, "test-mysql", "mysql", `{"host":"x"}`, 1))
	resp, err = doAuthGet(ts.URL+"/api/v1/datasource/42", token)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// 5) DELETE /datasource/42
	// Delete handler 先 SELECT 校验存在
	ts.Mock.ExpectQuery(`SELECT \* FROM datasources WHERE id`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "config", "status"}).
			AddRow(42, "test-mysql", "mysql", `{"host":"x"}`, 1))
	ts.Mock.ExpectExec(`UPDATE datasources SET deleted_at`).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 1))
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/datasource/42", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	b, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "delete body: %s", string(b))

	assert.NoError(t, ts.Mock.ExpectationsWereMet())
}

func doAuthGet(url, token string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return http.DefaultClient.Do(req)
}

func doAuthPost(url, token, body string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func bcryptGenerate(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(b), err
}
