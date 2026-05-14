import { describe, it, expect, vi } from 'vitest'

vi.mock('./request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import request from './request'
import { messageAPI } from './message'

describe('messageAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should list all messages', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await messageAPI.list()
    expect(request.get).toHaveBeenCalledWith('/messages')
  })

  it('should list unread messages only', async () => {
    vi.mocked(request.get).mockResolvedValue([])
    await messageAPI.list(true)
    expect(request.get).toHaveBeenCalledWith('/messages?unreadOnly=true')
  })

  it('should count unread messages', async () => {
    vi.mocked(request.get).mockResolvedValue(5)
    await messageAPI.countUnread()
    expect(request.get).toHaveBeenCalledWith('/messages/unread-count')
  })

  it('should mark message as read', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await messageAPI.markAsRead(1)
    expect(request.post).toHaveBeenCalledWith('/messages/1/read')
  })

  it('should mark all messages as read', async () => {
    vi.mocked(request.post).mockResolvedValue(undefined)
    await messageAPI.markAllAsRead()
    expect(request.post).toHaveBeenCalledWith('/messages/read-all')
  })

  it('should remove message', async () => {
    vi.mocked(request.delete).mockResolvedValue(undefined)
    await messageAPI.remove(1)
    expect(request.delete).toHaveBeenCalledWith('/messages/1')
  })
})
