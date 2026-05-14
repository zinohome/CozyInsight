import { renderHook, act, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useScreenData } from './useScreenData'
import type { Dashboard } from '@/types/dashboard'
import type { Chart, ChartDataResponse } from '@/types/chart'

// Mock chartAPI to avoid network calls
vi.mock('@/api/chart', () => ({
  chartAPI: {
    list: vi.fn(),
    getData: vi.fn(),
  },
}))

import { chartAPI } from '@/api/chart'

describe('useScreenData', () => {
  const mockDashboard: Dashboard = {
    id: 1,
    name: 'Screen 1',
    type: 'screen',
    config: JSON.stringify({
      canvas: { width: 1920, height: 1080, bgColor: '#000' },
      items: [
        { chartId: 1, positionX: 0, positionY: 0, width: 300, height: 200 },
        { chartId: 2, positionX: 400, positionY: 0, width: 300, height: 200 },
      ],
    }),
    createdBy: 1,
    createdAt: '2024-01-01',
    updatedAt: '2024-01-01',
  }

  const mockChartList: Chart[] = [
    { id: 1, title: 'Chart1', type: 'bar', datasetId: 1, config: '{}', status: 1, createdBy: 1, createdAt: '2024-01-01' },
    { id: 2, title: 'Chart2', type: 'line', datasetId: 1, config: '{}', status: 1, createdBy: 1, createdAt: '2024-01-01' },
  ]

  const mockData: ChartDataResponse = {
    dimensions: ['x'],
    metrics: ['y'],
    data: [{ x: 'a', y: 10 }],
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should initialize with default canvas and loading true', () => {
    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))
    expect(result.current.canvas).toEqual({ width: 1920, height: 1080, bgColor: '#0a1f44' })
    expect(result.current.loading).toBe(true)
    expect(result.current.items).toEqual([])
    expect(result.current.error).toBe('')
  })

  it('should load items from dashboard config on refetch', async () => {
    vi.mocked(chartAPI.list).mockResolvedValue(mockChartList)
    vi.mocked(chartAPI.getData).mockResolvedValue(mockData)

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.items).toHaveLength(2)
    expect(result.current.items[0].chart?.title).toBe('Chart1')
    expect(result.current.items[0].data).toEqual(mockData)
    expect(result.current.error).toBe('')
  })

  it('should show error for non-screen dashboard', async () => {
    const nonScreen: Dashboard = { ...mockDashboard, type: 'dashboard' }

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(nonScreen)))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.error).toBe('该资源不是数据大屏')
    expect(result.current.items).toEqual([])
  })

  it('should handle empty config gracefully', async () => {
    const emptyConfig: Dashboard = { ...mockDashboard, config: '{}' }

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(emptyConfig)))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.items).toEqual([])
  })

  it('should handle missing config gracefully', async () => {
    const noConfig: Dashboard = { ...mockDashboard, config: '' }

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(noConfig)))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.items).toEqual([])
  })

  it('should handle dashboard fetch error', async () => {
    const { result } = renderHook(() => useScreenData(() => Promise.reject(new Error('fail'))))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.error).toBe('加载数据大屏失败')
    expect(result.current.loading).toBe(false)
  })

  it('should set canvas from config', async () => {
    vi.mocked(chartAPI.list).mockResolvedValue(mockChartList)
    vi.mocked(chartAPI.getData).mockResolvedValue(mockData)

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.canvas).toEqual({
      width: 1920,
      height: 1080,
      bgColor: '#000',
    })
  })

  it('should cache chart list after first load', async () => {
    vi.mocked(chartAPI.list).mockResolvedValue(mockChartList)
    vi.mocked(chartAPI.getData).mockResolvedValue(mockData)

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))

    // Wait for mount polling to finish
    await waitFor(() => expect(result.current.items.length).toBe(2), { timeout: 2000 })
    expect(chartAPI.list).toHaveBeenCalledTimes(1)

    vi.clearAllMocks()

    // Trigger another refetch; list should not be called again
    await act(async () => {
      await result.current.refetch()
    })

    expect(chartAPI.list).not.toHaveBeenCalled()
    expect(result.current.items.length).toBe(2)
  })

  it('should handle per-chart data errors gracefully', async () => {
    vi.mocked(chartAPI.list).mockResolvedValue(mockChartList)
    // Let mount polling succeed first
    vi.mocked(chartAPI.getData).mockResolvedValue(mockData)

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))

    // Wait for mount polling to finish
    await waitFor(() => expect(result.current.items.length).toBe(2), { timeout: 2000 })

    // Now set up a mixed mock for the next refetch
    vi.clearAllMocks()
    vi.mocked(chartAPI.getData)
      .mockResolvedValueOnce(mockData)
      .mockRejectedValueOnce(new Error('chart error'))

    await act(async () => {
      await result.current.refetch()
    })

    expect(result.current.items).toHaveLength(2)
    expect(result.current.items[0].data).toBeDefined()
    expect(result.current.items[1].data).toBeUndefined()
  })

  it('should calculate scale on mount', async () => {
    vi.mocked(chartAPI.list).mockResolvedValue(mockChartList)
    vi.mocked(chartAPI.getData).mockResolvedValue(mockData)

    const { result } = renderHook(() => useScreenData(() => Promise.resolve(mockDashboard)))

    await act(async () => {
      await result.current.refetch()
    })

    const expectedScale = Math.min(window.innerWidth / 1920, window.innerHeight / 1080)
    expect(result.current.scale).toBe(expectedScale)
  })
})
