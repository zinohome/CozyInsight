export const exportAPI = {
  downloadCSV: (chartId: number) => {
    window.open(`/api/v1/chart/${chartId}/export/csv`, '_blank')
  },
}
