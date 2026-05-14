import { describe, it, expect, vi, beforeEach } from 'vitest'
import { exportAPI } from './export'

describe('exportAPI', () => {
  let openSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    openSpy = vi.spyOn(window, 'open').mockImplementation(() => null)
  })

  it('should open CSV export in new window', () => {
    exportAPI.downloadCSV(1)
    expect(openSpy).toHaveBeenCalledWith('/api/v1/chart/1/export/csv', '_blank')
  })

  it('should open Excel export in new window', () => {
    exportAPI.downloadExcel(1)
    expect(openSpy).toHaveBeenCalledWith('/api/v1/chart/1/export/excel', '_blank')
  })
})
