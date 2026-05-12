export const exportAPI = {
  downloadCSV: (chartId: number) => {
    window.open(`/api/v1/chart/${chartId}/export/csv`, '_blank')
  },
  downloadExcel: (chartId: number) => {
    window.open(`/api/v1/chart/${chartId}/export/excel`, '_blank')
  },
}
