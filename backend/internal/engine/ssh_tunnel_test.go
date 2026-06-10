package engine

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSSHDialer 替换 NewSSHTunnel 的真实 ssh.Dial 步骤过于侵入，所以仅在解析层测试。
// 这里聚焦在参数解析失败路径上以提升覆盖率。

func TestNewSSHTunnel_InvalidPrivateKey(t *testing.T) {
	_, err := NewSSHTunnel(
		"unused", 22, "user", "",
		"-----BEGIN INVALID KEY-----\nnot-a-key\n",
		"db.internal", 3306, 0,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse private key")
}

func TestNewSSHTunnel_NoAuth(t *testing.T) {
	_, err := NewSSHTunnel(
		"unused", 22, "user", "", "",
		"db.internal", 3306, 0,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth method")
}

// TestSSHTunnel_LocalPort_ListenerNil 覆盖 LocalPort 在 listener==nil 分支
// （典型发生在隧道尚未启动时通过 NewSSHTunnel 失败后）。
func TestSSHTunnel_LocalPort_ListenerNil(t *testing.T) {
	tn := &SSHTunnel{localPort: 12345}
	assert.Equal(t, 12345, tn.LocalPort())
	assert.Equal(t, "127.0.0.1:12345", tn.LocalAddr())
}

// TestSSHTunnel_Close_NoListener 覆盖 Close 时的 nil-safety。
func TestSSHTunnel_Close_NoListener(t *testing.T) {
	tn := &SSHTunnel{}
	assert.NoError(t, tn.Close())
}

// TestSSHTunnel_Close_WithListener 覆盖 listener 已打开但 sshClient 为 nil 的情况。
func TestSSHTunnel_Close_WithListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	tn := &SSHTunnel{listener: ln}
	assert.NoError(t, tn.Close())
}
