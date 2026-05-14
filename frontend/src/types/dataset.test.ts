import { describe, it, expect } from 'vitest'
import type { Dataset, DatasetField, CreateDatasetRequest, PreviewDataResponse } from './dataset'

describe('dataset types', () => {
  it('should allow valid Dataset', () => {
    const ds: Dataset = {
      id: 1,
      name: 'Sales Data',
      datasourceId: 1,
      databaseName: 'sales_db',
      tableName: 'orders',
      sql: 'SELECT * FROM orders',
      type: 'db',
      mode: 0,
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(ds.tableName).toBe('orders')
  })

  it('should allow Dataset without optional sql', () => {
    const ds: Dataset = {
      id: 1,
      name: 'Sales Data',
      datasourceId: 1,
      databaseName: 'sales_db',
      tableName: 'orders',
      type: 'db',
      mode: 0,
      status: 1,
      createdBy: 1,
      createdAt: '2024-01-01',
    }
    expect(ds.sql).toBeUndefined()
  })

  it('should allow valid DatasetField', () => {
    const field: DatasetField = {
      id: 1,
      datasetId: 1,
      name: 'amount',
      type: 'DECIMAL',
      deType: 2,
      length: 18,
      precision: 18,
      scale: 2,
      originName: 'amount',
    }
    expect(field.deType).toBe(2)
  })

  it('should allow valid CreateDatasetRequest', () => {
    const req: CreateDatasetRequest = {
      name: 'New Dataset',
      datasourceId: 1,
      databaseName: 'db1',
      tableName: 't1',
      type: 'db',
      mode: 0,
    }
    expect(req.mode).toBe(0)
  })

  it('should allow valid PreviewDataResponse', () => {
    const resp: PreviewDataResponse = {
      fields: [
        { id: 1, name: 'id', type: 'INT', deType: 0, length: 11, precision: 0, scale: 0, originName: 'id' },
      ],
      data: [{ id: 1, name: 'Test' }],
    }
    expect(resp.fields).toHaveLength(1)
    expect(resp.data).toHaveLength(1)
  })
})
