import { describe, it, expect } from 'vitest'
import type { Message } from './message'

describe('message types', () => {
  it('should allow valid unread Message', () => {
    const msg: Message = {
      id: 1,
      userId: 1,
      title: 'New notification',
      content: 'You have a new message',
      type: 'system',
      isRead: 0,
      createdAt: '2024-01-01',
    }
    expect(msg.isRead).toBe(0)
    expect(msg.type).toBe('system')
  })

  it('should allow valid read Message', () => {
    const msg: Message = {
      id: 2,
      userId: 1,
      title: 'Old notification',
      content: 'Read message',
      type: 'alert',
      isRead: 1,
      createdAt: '2024-01-01',
    }
    expect(msg.isRead).toBe(1)
  })
})
