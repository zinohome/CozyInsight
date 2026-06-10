//go:build integration
// +build integration

package container

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"cozy-insight/internal/engine"
)

// TestSSHTunnel_RealContainer 真实 sshd 容器集成测试。
//
// 架构：sshd 容器内跑一个 netcat echo server；用 SSHTunnel 把容器内的 echo 端口
// 转发到 host 的本地回环端口；host 端用普通 TCP 连过去发/收数据，验证 SSH 隧道
// 端到端可用。
//
// 这个测试覆盖了：
//   - SSH 客户端握手（密码认证）
//   - SSHTunnel 的本地 listen
//   - 通过 SSH 直接 TCP 通道转发的 forward
//   - 双向 io.Copy 转发数据
func TestSSHTunnel_RealContainer(t *testing.T) {
	requireDocker(t)
	if testing.Short() {
		t.Skip("short mode")
	}

	ctx := context.Background()
	cmds := `
apk add --no-cache openssh-server openssh-keygen busybox-extras >/dev/null 2>&1
ssh-keygen -A >/dev/null 2>&1
printf 'PermitRootLogin yes\nPasswordAuthentication yes\nAllowTcpForwarding yes\nUsePAM no\n' > /etc/ssh/sshd_config.d/permit.conf
echo 'root:testpass' | chpasswd
# 在容器里跑一个 netcat echo server，监听 2223
nohup nc -lk -p 2223 -e /bin/cat >/dev/null 2>&1 &
exec /usr/sbin/sshd -D -e -p 2222
`
	sshReq := testcontainers.ContainerRequest{
		Image:        "alpine:3.20",
		ExposedPorts: []string{"2222/tcp"},
		Cmd:          []string{"/bin/sh", "-c", cmds},
		WaitingFor:   wait.ForLog("Server listening on").WithStartupTimeout(120 * time.Second),
	}
	sshCtr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: sshReq,
		Started:          true,
	})
	require.NoError(t, err, "start sshd container")
	defer func() { _ = sshCtr.Terminate(context.Background()) }()

	sshHost, err := sshCtr.Host(ctx)
	require.NoError(t, err)
	sshPort, err := sshCtr.MappedPort(ctx, "2222/tcp")
	require.NoError(t, err)
	sshPortInt, err := strconv.Atoi(sshPort.Port())
	require.NoError(t, err)
	// Force 127.0.0.1 on Mac (IPv6 resolution race).
	sshHost = "127.0.0.1"

	// SSH 隧道：把容器内 127.0.0.1:2223（netcat echo）转发到 host 的本地回环端口。
	tunnel, err := engine.NewSSHTunnel(
		sshHost, sshPortInt,
		"root", "testpass", "",
		"127.0.0.1", 2223, 0, // 0 = 系统分配
	)
	require.NoError(t, err, "build ssh tunnel")
	defer func() { _ = tunnel.Close() }()

	t.Logf("ssh tunnel: ssh=%s:%d remote=127.0.0.1:2223 local=%s",
		sshHost, sshPortInt, tunnel.LocalAddr())

	// 通过隧道发/收
	conn, err := net.DialTimeout("tcp", tunnel.LocalAddr(), 5*time.Second)
	require.NoError(t, err, "dial tunnel local")
	defer conn.Close()

	msg := []byte("hello-tunnel")
	_, err = conn.Write(msg)
	require.NoError(t, err)
	buf := make([]byte, len(msg))
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err = conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, msg, buf)
}
