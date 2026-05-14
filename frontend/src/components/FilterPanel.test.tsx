import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import FilterPanel from './FilterPanel'

describe('FilterPanel', () => {
  const fields = ['name', 'age', 'city']

  it('should render empty filters with add button', () => {
    render(
      <FilterPanel fields={fields} filters={[]} onChange={vi.fn()} />
    )
    expect(screen.getByText('筛选条件:')).toBeInTheDocument()
    expect(screen.getByText('添加筛选')).toBeInTheDocument()
  })

  it('should add new filter when clicking add button', () => {
    const onChange = vi.fn()
    render(
      <FilterPanel fields={fields} filters={[]} onChange={onChange} />
    )

    fireEvent.click(screen.getByText('添加筛选'))
    expect(onChange).toHaveBeenCalledTimes(1)
    expect(onChange.mock.calls[0][0]).toHaveLength(1)
    expect(onChange.mock.calls[0][0][0].field).toBe('name')
    expect(onChange.mock.calls[0][0][0].operator).toBe('=')
  })

  it('should render existing filters', () => {
    const filters = [
      { id: 'f1', field: 'name', operator: '=', value: 'test' },
    ]
    render(
      <FilterPanel fields={fields} filters={filters} onChange={vi.fn()} />
    )
    expect(screen.getByText('name')).toBeInTheDocument()
  })

  it('should update filter value through input change', () => {
    const onChange = vi.fn()
    const filters = [{ id: 'f1', field: 'name', operator: '=', value: '' }]
    render(
      <FilterPanel fields={fields} filters={filters} onChange={onChange} />
    )

    const input = screen.getByPlaceholderText('值')
    fireEvent.change(input, { target: { value: 'new value' } })

    expect(onChange).toHaveBeenCalled()
    expect(onChange.mock.calls[0][0][0].value).toBe('new value')
  })

  it('should update filter field through select change', async () => {
    const onChange = vi.fn()
    const filters = [{ id: 'f1', field: 'name', operator: '=', value: '' }]
    const { container } = render(
      <FilterPanel fields={fields} filters={filters} onChange={onChange} />
    )

    // Find the first ant-select (field selector) and open it
    const selects = container.querySelectorAll('.ant-select')
    fireEvent.mouseDown(selects[0])

    // Wait for dropdown and click the 'age' option
    await waitFor(() => {
      const option = document.querySelector('.ant-select-dropdown')?.querySelector('[title="age"]')
      if (option) fireEvent.click(option)
    })

    expect(onChange).toHaveBeenCalled()
  })

  it('should update filter operator through select change', async () => {
    const onChange = vi.fn()
    const filters = [{ id: 'f1', field: 'name', operator: '=', value: '' }]
    const { container } = render(
      <FilterPanel fields={fields} filters={filters} onChange={onChange} />
    )

    // Find the second ant-select (operator selector) and open it
    const selects = container.querySelectorAll('.ant-select')
    fireEvent.mouseDown(selects[1])

    // Wait for dropdown and click the '!=' option
    await waitFor(() => {
      const option = document.querySelector('.ant-select-dropdown')?.querySelector('[title="不等于"]')
      if (option) fireEvent.click(option)
    })

    expect(onChange).toHaveBeenCalled()
  })

  it('should remove filter when clicking close', () => {
    const onChange = vi.fn()
    const filters = [{ id: 'f1', field: 'name', operator: '=', value: 'test' }]
    render(
      <FilterPanel fields={fields} filters={filters} onChange={onChange} />
    )

    const closeBtn = screen.getByRole('img', { name: /close/i })
    fireEvent.click(closeBtn)
    expect(onChange).toHaveBeenCalledWith([])
  })

  it('should handle empty fields array', () => {
    const onChange = vi.fn()
    render(
      <FilterPanel fields={[]} filters={[]} onChange={onChange} />
    )

    fireEvent.click(screen.getByText('添加筛选'))
    expect(onChange).toHaveBeenCalledTimes(1)
    expect(onChange.mock.calls[0][0][0].field).toBe('')
  })

  it('should render multiple filters', () => {
    const filters = [
      { id: 'f1', field: 'name', operator: '=', value: 'a' },
      { id: 'f2', field: 'age', operator: '>', value: '18' },
    ]
    const { container } = render(
      <FilterPanel fields={fields} filters={filters} onChange={vi.fn()} />
    )
    expect(container.textContent).toContain('name')
    expect(container.textContent).toContain('age')
  })

  it('should use first field as default when adding filter', () => {
    const onChange = vi.fn()
    render(
      <FilterPanel fields={['first']} filters={[]} onChange={onChange} />
    )
    fireEvent.click(screen.getByText('添加筛选'))
    expect(onChange.mock.calls[0][0][0].field).toBe('first')
  })
})
