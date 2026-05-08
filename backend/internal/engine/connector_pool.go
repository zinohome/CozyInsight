package engine

import (
	"fmt"
	"sync"
)

// ConnectorPool caches DatasourceConnector instances keyed by datasource ID.
// It is safe for concurrent use.
type ConnectorPool struct {
	mu    sync.RWMutex
	pools map[uint64]DatasourceConnector
}

// NewConnectorPool creates a new empty connector pool.
func NewConnectorPool() *ConnectorPool {
	return &ConnectorPool{pools: make(map[uint64]DatasourceConnector)}
}

// Get returns an existing connector for the given datasource ID, or creates
// and connects a new one if none exists. The connector is retained in the
// pool for subsequent calls.
func (p *ConnectorPool) Get(dsID uint64, dsType string, configJSON string) (DatasourceConnector, error) {
	p.mu.RLock()
	conn, ok := p.pools[dsID]
	p.mu.RUnlock()
	if ok {
		return conn, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	conn, ok = p.pools[dsID]
	if ok {
		return conn, nil
	}

	conn, err := NewConnector(dsType)
	if err != nil {
		return nil, fmt.Errorf("create connector failed: %w", err)
	}
	if err := conn.Connect(configJSON); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	p.pools[dsID] = conn
	return conn, nil
}

// Remove disconnects and removes the connector for the given datasource ID.
func (p *ConnectorPool) Remove(dsID uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if conn, ok := p.pools[dsID]; ok {
		conn.Close()
		delete(p.pools, dsID)
	}
}

// Close disconnects all pooled connectors and clears the pool.
func (p *ConnectorPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, conn := range p.pools {
		conn.Close()
	}
	p.pools = make(map[uint64]DatasourceConnector)
}
