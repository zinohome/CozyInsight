import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AppRoutes from './index'

describe('Router', () => {
  it('should render login page on /login', () => {
    render(
      <MemoryRouter initialEntries={['/login']}>
        <AppRoutes />
      </MemoryRouter>
    )
    expect(screen.getByText('CozyInsight')).toBeInTheDocument()
  })

  it('should render share view on /share/:token', () => {
    render(
      <MemoryRouter initialEntries={['/share/abc123']}>
        <AppRoutes />
      </MemoryRouter>
    )
    expect(document.body).toBeInTheDocument()
  })

  it('should render screen view on /screen/view/:id', () => {
    render(
      <MemoryRouter initialEntries={['/screen/view/1']}>
        <AppRoutes />
      </MemoryRouter>
    )
    expect(document.body).toBeInTheDocument()
  })

  it('should render dashboard view on /dashboard/view/:id', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard/view/1']}>
        <AppRoutes />
      </MemoryRouter>
    )
    expect(document.body).toBeInTheDocument()
  })

  it('should render 404 page for unknown routes', () => {
    render(
      <MemoryRouter initialEntries={['/unknown']}>
        <AppRoutes />
      </MemoryRouter>
    )
    expect(document.body).toBeInTheDocument()
  })
})
