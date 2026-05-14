import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import NotFoundPage from './index'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('NotFoundPage', () => {
  it('should render 404 status and title', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>
    )
    expect(screen.getByText('404')).toBeInTheDocument()
  })

  it('should render sub title', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>
    )
    expect(screen.getByText('抱歉，您访问的页面不存在。')).toBeInTheDocument()
  })

  it('should render back button', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>
    )
    expect(screen.getByText('返回工作台')).toBeInTheDocument()
  })

  it('should navigate to home when clicking back button', () => {
    render(
      <MemoryRouter>
        <NotFoundPage />
      </MemoryRouter>
    )
    const backBtn = screen.getByText('返回工作台')
    fireEvent.click(backBtn)
    expect(mockNavigate).toHaveBeenCalledWith('/')
  })
})
