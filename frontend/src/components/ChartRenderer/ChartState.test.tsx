import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ChartEmpty, ChartLoading, ChartError } from './ChartState'

describe('ChartState', () => {
  describe('ChartEmpty', () => {
    it('renders default empty message', () => {
      render(<ChartEmpty />)
      expect(screen.getByText('暂无数据')).toBeInTheDocument()
    })

    it('renders custom message', () => {
      render(<ChartEmpty message="没有匹配的数据" />)
      expect(screen.getByText('没有匹配的数据')).toBeInTheDocument()
    })

    it('respects custom height', () => {
      const { container } = render(<ChartEmpty height={500} />)
      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.style.height).toBe('500px')
    })
  })

  describe('ChartLoading', () => {
    it('renders loading state', () => {
      const { container } = render(<ChartLoading />)
      expect(container.firstChild).toBeInTheDocument()
    })

    it('respects custom height', () => {
      const { container } = render(<ChartLoading height={400} />)
      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.style.height).toBe('400px')
    })
  })

  describe('ChartError', () => {
    it('renders error message', () => {
      render(<ChartError error="数据库连接超时" />)
      expect(screen.getByText('加载失败')).toBeInTheDocument()
      expect(screen.getByText('数据库连接超时')).toBeInTheDocument()
    })

    it('shows retry button when onRetry provided', () => {
      const onRetry = vi.fn()
      render(<ChartError error="网络异常" onRetry={onRetry} />)
      const btn = screen.getByRole('button', { name: /重试/ })
      expect(btn).toBeInTheDocument()
      fireEvent.click(btn)
      expect(onRetry).toHaveBeenCalledTimes(1)
    })

    it('hides retry button when onRetry not provided', () => {
      render(<ChartError error="权限不足" />)
      expect(screen.queryByRole('button', { name: /重试/ })).not.toBeInTheDocument()
    })

    it('respects custom height', () => {
      const { container } = render(<ChartError error="x" height={250} />)
      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.style.height).toBe('250px')
    })
  })
})
