import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

// Stub antd message globally (it uses static config hooks that hang in jsdom).
vi.mock('antd', async () => {
  const actual = await vi.importActual<any>('antd')
  return {
    ...actual,
    message: { success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn() },
  }
})

vi.mock('@/api/user', () => ({
  userAPI: {
    update: vi.fn(),
    profile: vi.fn(),
    changePassword: vi.fn(),
  },
}))

const mockSetUser = vi.fn()
const authState = {
  user: { id: 1, username: 'alice', email: 'a@x.com', nickName: 'Alice', isAdmin: false },
  setUser: mockSetUser,
}
vi.mock('@/store/auth', () => ({
  useAuthStore: Object.assign(
    vi.fn((selector: (s: any) => any) => selector(authState)),
    { getState: () => authState }
  ),
}))

// Import AFTER mocks
import ProfilePage from './index'

describe('ProfilePage smoke', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )
    // 基本信息 card title and 修改密码 card title are present
    expect(screen.getByText('基本信息')).toBeInTheDocument()
    expect(screen.getAllByText('修改密码').length).toBeGreaterThanOrEqual(1)
  })

  it('shows user nickname in profile header', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )
    expect(screen.getByText('Alice')).toBeInTheDocument()
  })

  it('shows "普通用户" for non-admin user', () => {
    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>
    )
    expect(screen.getByText('普通用户')).toBeInTheDocument()
  })
})
