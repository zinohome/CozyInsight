import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import DrillBreadcrumb from './DrillBreadcrumb'

describe('DrillBreadcrumb', () => {
  it('should render nothing when only one dimension', () => {
    const { container } = render(
      <DrillBreadcrumb dimensions={['province']} currentLevel={0} onDrillUp={vi.fn()} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('should render dimensions up to current level', () => {
    render(
      <DrillBreadcrumb
        dimensions={['province', 'city', 'district']}
        currentLevel={1}
        onDrillUp={vi.fn()}
      />
    )
    expect(screen.getByText('province')).toBeInTheDocument()
    expect(screen.getByText('city')).toBeInTheDocument()
    expect(screen.queryByText('district')).not.toBeInTheDocument()
  })

  it('should make previous dimensions clickable', () => {
    const onDrillUp = vi.fn()
    render(
      <DrillBreadcrumb
        dimensions={['province', 'city', 'district']}
        currentLevel={2}
        onDrillUp={onDrillUp}
      />
    )

    // Province should be a clickable button (not current level)
    const provinceBtn = screen.getByRole('button', { name: 'province' })
    fireEvent.click(provinceBtn)
    expect(onDrillUp).toHaveBeenCalledWith(0)
  })

  it('should not make current dimension clickable', () => {
    const onDrillUp = vi.fn()
    render(
      <DrillBreadcrumb
        dimensions={['province', 'city']}
        currentLevel={1}
        onDrillUp={onDrillUp}
      />
    )

    // City is current level, should be text not button
    expect(screen.getByText('city')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'city' })).not.toBeInTheDocument()
  })

  it('should render all dimensions at max level', () => {
    render(
      <DrillBreadcrumb
        dimensions={['province', 'city', 'district']}
        currentLevel={2}
        onDrillUp={vi.fn()}
      />
    )
    expect(screen.getByText('province')).toBeInTheDocument()
    expect(screen.getByText('city')).toBeInTheDocument()
    expect(screen.getByText('district')).toBeInTheDocument()
  })
})
