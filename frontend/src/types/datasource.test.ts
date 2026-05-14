import { describe, it, expect } from 'vitest'
import type { Datasource, CreateDatasourceRequest, TestConnectionRequest } from './datasource'

describe('datasource types', () => {
  it('should allow valid Datasource', () => {
    const ds: Datasource = {
      id: 1,
      name: 'MySQL Production',
      type: 'mysql',
      config: JSON.stringify({ host: 'localhost', port: 3306 }),
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(ds.type).toBe('mysql')
  })

  it('should allow Datasource with file fields', () => {
    const ds: Datasource = {
      id: 2,
      name: 'Sales Excel',
      type: 'excel',
      config: '{}',
      filePath: '/uploads/sales.xlsx',
      fileType: 'xlsx',
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(ds.fileType).toBe('xlsx')
  })

  it('should allow valid CreateDatasourceRequest', () => {
    const req: CreateDatasourceRequest = {
      name: 'New DB',
      type: 'postgresql',
      config: '{}',
    }
    expect(req.type).toBe('postgresql')
  })

  it('should allow valid TestConnectionRequest', () => {
    const req: TestConnectionRequest = {
      type: 'mysql',
      config: { host: 'localhost', port: 3306, database: 'test' },
    }
    expect(req.config.host).toBe('localhost')
  })
})
