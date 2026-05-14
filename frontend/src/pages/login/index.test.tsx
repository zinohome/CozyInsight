import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from './index'

vi.mock('@/api/auth', () => ({
  authAPI: {
    login: vi.fn(),
    register: vi.fn(),
  },
}))

vi.mock('@/store/auth', () => ({
  useAuthStore: vi.fn(() => ({
    setToken: vi.fn(),
  })),
}))

describe('LoginPage', () => {
  const renderLogin = () => {
    return render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render login form with required fields', () => {
    renderLogin()
    expect(screen.getByPlaceholderText('用户名')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('密码')).toBeInTheDocument()
    expect(screen.getByText('登录')).toBeInTheDocument()
  })

  it('should render card title', () => {
    renderLogin()
    expect(screen.getByText('CozyInsight')).toBeInTheDocument()
  })

  it('should have login and register tabs', () => {
    renderLogin()
    expect(screen.getByText('登录')).toBeInTheDocument()
    expect(screen.getByText('注册')).toBeInTheDocument()
  })

  it('should switch to register tab', async () => {
    renderLogin()
    const registerTab = screen.getByText('注册')
    await act(async () => {
      fireEvent.click(registerTab)
    })
    expect(screen.getByPlaceholderText('邮箱')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('昵称（可选）')).toBeInTheDocument()
  })

  it('should render center-aligned container', () => {
    const { container } = renderLogin()
    const wrapper = container.querySelector('div[style*="flex"]')
    expect(wrapper).toBeInTheDocument()
  })
})
