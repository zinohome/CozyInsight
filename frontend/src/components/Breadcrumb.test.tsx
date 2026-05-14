import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AppBreadcrumb from './Breadcrumb'

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

describe('AppBreadcrumb', () => {
  const renderWithRouter = (path: string) => {
    return render(
      <MemoryRouter initialEntries={[path]}>
        <AppBreadcrumb />
      </MemoryRouter>
    )
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render home breadcrumb on root path', () => {
    renderWithRouter('/')
    expect(screen.getByText('工作台')).toBeInTheDocument()
  })

  it('should render breadcrumb for datasource page', () => {
    renderWithRouter('/datasource')
    expect(screen.getByText('工作台')).toBeInTheDocument()
    expect(screen.getByText('数据源')).toBeInTheDocument()
  })

  it('should render breadcrumb for chart page', () => {
    renderWithRouter('/chart')
    expect(screen.getByText('图表')).toBeInTheDocument()
  })

  it('should render breadcrumb for dashboard page', () => {
    renderWithRouter('/dashboard')
    expect(screen.getByText('仪表板')).toBeInTheDocument()
  })

  it('should render breadcrumb for system user page', () => {
    renderWithRouter('/system/user')
    expect(screen.getByText('工作台')).toBeInTheDocument()
    expect(screen.getByText('system')).toBeInTheDocument()
    expect(screen.getByText('用户管理')).toBeInTheDocument()
  })

  it('should render breadcrumb for chart builder with dynamic id', () => {
    renderWithRouter('/chart/builder/123')
    expect(screen.getByText('图表')).toBeInTheDocument()
    expect(screen.getByText('图表设计器')).toBeInTheDocument()
    expect(screen.getByText('123')).toBeInTheDocument()
  })

  it('should make home link clickable', () => {
    renderWithRouter('/dataset')
    const homeLink = screen.getByText('工作台')
    expect(homeLink.tagName).toBe('A')
  })

  it('should navigate on home click', () => {
    renderWithRouter('/dataset')
    const homeLink = screen.getByText('工作台')
    fireEvent.click(homeLink)
    expect(mockNavigate).toHaveBeenCalledWith('/')
  })

  it('should navigate on breadcrumb item click', () => {
    renderWithRouter('/datasource')
    const link = screen.getByText('数据源')
    fireEvent.click(link)
    expect(mockNavigate).toHaveBeenCalledWith('/datasource')
  })

  it('should render unknown path using segment name', () => {
    renderWithRouter('/unknown/path')
    expect(screen.getByText('工作台')).toBeInTheDocument()
  })
})
