import { describe, it, expect } from 'vitest'
import type { OperationLog } from './log'

describe('log types', () => {
  it('should allow valid OperationLog', () => {
    const log: OperationLog = {
      id: 1,
      userId: 1,
      username: 'admin',
      method: 'POST',
      path: '/api/v1/datasource',
      query: '',
      body: '{"name":"test"}',
      ip: '127.0.0.1',
      userAgent: 'Mozilla/5.0',
      statusCode: 200,
      duration: 150,
      errorMessage: '',
      createdAt: '2024-01-01',
    }
    expect(log.method).toBe('POST')
    expect(log.statusCode).toBe(200)
    expect(log.duration).toBe(150)
  })

  it('should allow OperationLog with error', () => {
    const log: OperationLog = {
      id: 2,
      userId: 1,
      username: 'admin',
      method: 'DELETE',
      path: '/api/v1/datasource/1',
      query: '',
      body: '',
      ip: '127.0.0.1',
      userAgent: 'Mozilla/5.0',
      statusCode: 500,
      duration: 50,
      errorMessage: 'Connection refused',
      createdAt: '2024-01-01',
    }
    expect(log.statusCode).toBe(500)
    expect(log.errorMessage).toBe('Connection refused')
  })
})
