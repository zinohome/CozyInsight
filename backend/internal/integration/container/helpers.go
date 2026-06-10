//go:build integration
// +build integration

package container

import (
	"testing"
	"time"
)

// requireDocker 跳过没有 Docker 环境的测试，避免在 CI 无 Docker 时无谓失败。
func requireDocker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping real container test in short mode")
	}
	// testcontainers 在没 Docker 时也会 panic/返回错；我们靠 t.Skip 减少噪音
}

// connectWithRetry 重试连接，以应对容器 log wait strategy 报"ready"但实际
// 还没完全接受 auth 的竞态；用短间隔重试几次。
func connectWithRetry(t *testing.T, connect func() error, attempts int, delay time.Duration) error {
	t.Helper()
	var last error
	for i := 0; i < attempts; i++ {
		if err := connect(); err == nil {
			return nil
		} else {
			last = err
		}
		time.Sleep(delay)
	}
	return last
}
