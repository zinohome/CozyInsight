//go:build integration
// +build integration

// Package container 提供了基于 testcontainers 的真实数据库容器集成测试。
//
// 这些测试默认不参与 `go test ./...`；需要显式开启：
//
//	go test -tags=integration -timeout 5m ./internal/integration/container/...
//
// 跳过策略：CI 跑时如果有 Docker 可用，集成测试会自动运行。
// 跳过：在没有 Docker 环境下，可以加 -short 让 `testing.Short()` 跳过。
package container
