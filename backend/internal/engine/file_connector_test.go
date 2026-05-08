package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestFileConnector_Connect_InvalidConfig(t *testing.T) {
	c := &fileConnector{}
	err := c.Connect("invalid json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config json")
}

func TestFileConnector_Connect_MissingFilePath(t *testing.T) {
	c := &fileConnector{}
	err := c.Connect(`{"file_type":"csv"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing or invalid file_path")
}

func TestFileConnector_Connect_UnsupportedType(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(tmp, []byte("hello"), 0644))

	c := &fileConnector{}
	err := c.Connect(`{"file_path":"` + tmp + `","file_type":"txt"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestFileConnector_CSV_QueryAndColumns(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.csv")
	require.NoError(t, os.WriteFile(tmp, []byte("name,age\nAlice,30\nBob,25\n"), 0644))

	c := &fileConnector{}
	require.NoError(t, c.Connect(`{"file_path":"`+tmp+`","file_type":"csv"}`))
	defer c.Close()

	cols, err := c.GetColumns(context.Background(), "", "data")
	require.NoError(t, err)
	require.Len(t, cols, 2)
	assert.Equal(t, "name", cols[0].Name)
	assert.Equal(t, "age", cols[1].Name)

	rows, err := c.Query(context.Background(), "SELECT * FROM data WHERE age > ?", 25)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Alice", rows[0]["name"])
}

func TestFileConnector_Excel_QueryAndColumns(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.xlsx")
	f := excelize.NewFile()
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "product"))
	require.NoError(t, f.SetCellValue("Sheet1", "B1", "sales"))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", "A"))
	require.NoError(t, f.SetCellValue("Sheet1", "B2", 100))
	require.NoError(t, f.SetCellValue("Sheet1", "A3", "B"))
	require.NoError(t, f.SetCellValue("Sheet1", "B3", 200))
	require.NoError(t, f.SaveAs(tmp))
	f.Close()

	c := &fileConnector{}
	require.NoError(t, c.Connect(`{"file_path":"`+tmp+`","file_type":"excel"}`))
	defer c.Close()

	cols, err := c.GetColumns(context.Background(), "", "data")
	require.NoError(t, err)
	require.Len(t, cols, 2)
	assert.Equal(t, "product", cols[0].Name)
	assert.Equal(t, "sales", cols[1].Name)

	rows, err := c.Query(context.Background(), "SELECT * FROM data")
	require.NoError(t, err)
	require.Len(t, rows, 2)
}

func TestFileConnector_Query_NotConnected(t *testing.T) {
	c := &fileConnector{}
	_, err := c.Query(context.Background(), "SELECT 1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFileConnector_GetColumns_NotConnected(t *testing.T) {
	c := &fileConnector{}
	_, err := c.GetColumns(context.Background(), "", "data")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestFileConnector_Connect_Idempotent(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.csv")
	require.NoError(t, os.WriteFile(tmp, []byte("a\n1\n"), 0644))

	c := &fileConnector{}
	require.NoError(t, c.Connect(`{"file_path":"`+tmp+`","file_type":"csv"}`))
	require.NoError(t, c.Connect(`{"file_path":"`+tmp+`","file_type":"csv"}`))
	defer c.Close()
	assert.NotNil(t, c.db)
}

func TestFileConnector_Close_Idempotent(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.csv")
	require.NoError(t, os.WriteFile(tmp, []byte("a\n1\n"), 0644))

	c := &fileConnector{}
	require.NoError(t, c.Connect(`{"file_path":"`+tmp+`","file_type":"csv"}`))
	require.NoError(t, c.Close())
	require.NoError(t, c.Close())
}

func TestFileConnector_EmptyCSV(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "empty.csv")
	require.NoError(t, os.WriteFile(tmp, []byte(""), 0644))

	c := &fileConnector{}
	err := c.Connect(`{"file_path":"` + tmp + `","file_type":"csv"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty csv file")
}

func TestFileConnector_EmptyExcel(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "empty.xlsx")
	f := excelize.NewFile()
	require.NoError(t, f.SaveAs(tmp))
	f.Close()

	c := &fileConnector{}
	err := c.Connect(`{"file_path":"` + tmp + `","file_type":"excel"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty excel file")
}

func TestSanitizeName(t *testing.T) {
	assert.Equal(t, "hello_world", sanitizeName("hello world", 0))
	assert.Equal(t, "hello_world", sanitizeName("hello-world", 1))
	assert.Equal(t, "foo", sanitizeName("  foo  ", 2))
	assert.Equal(t, "col0", sanitizeName("", 0))
	assert.Equal(t, "col1", sanitizeName("", 1))
}
