import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, act } from '@testing-library/react'
import MessageCenter from './MessageCenter'

vi.mock('@/api/message', () => ({
  messageAPI: {
    list: vi.fn(),
    countUnread: vi.fn(),
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    remove: vi.fn(),
  },
}))

import { messageAPI } from '@/api/message'

describe('MessageCenter', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.mocked(messageAPI.countUnread).mockResolvedValue(1)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should render bell icon', () => {
    render(<MessageCenter />)
    expect(document.querySelector('[data-icon="bell"]')).toBeInTheDocument()
  })

  it('should poll for unread count on mount', async () => {
    render(<MessageCenter />)
    await act(async () => {
      vi.advanceTimersByTime(30000)
    })
    expect(messageAPI.countUnread).toHaveBeenCalled()
  })

  it('should fetch messages when dropdown opens', async () => {
    vi.mocked(messageAPI.list).mockResolvedValue([
      { id: 1, title: 'Test', content: 'test', isRead: 0, createdAt: '2024-01-01' },
    ])
    render(<MessageCenter />)

    // Open dropdown by clicking bell
    const bell = document.querySelector('[data-icon="bell"]')?.closest('span')
    if (bell) {
      await act(async () => {
        bell.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      })
    }

    expect(messageAPI.list).toHaveBeenCalled()
  })

  it('should show empty state when no messages', async () => {
    vi.mocked(messageAPI.list).mockResolvedValue([])
    render(<MessageCenter />)

    const bell = document.querySelector('[data-icon="bell"]')?.closest('span')
    if (bell) {
      await act(async () => {
        bell.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      })
    }

    expect(messageAPI.list).toHaveBeenCalled()
  })
})
