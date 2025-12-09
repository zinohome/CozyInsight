import request from '../utils/request';

export const exportAPI = {
    // 导出数据集
    exportDataset: (id: string, format: 'excel' | 'csv' = 'excel') => {
        const url = `/dataset/${id}/export?format=${format}`;
        return request.get(url, { responseType: 'blob' }).then((blob) => {
            const filename = `dataset_${id}.${format === 'csv' ? 'csv' : 'xlsx'}`;
            downloadBlob(blob, filename);
        });
    },

    // 导出图表数据
    exportChartData: (id: string, format: 'excel' | 'csv' = 'excel') => {
        const url = `/chart/${id}/export?format=${format}`;
        return request.get(url, { responseType: 'blob' }).then((blob) => {
            const filename = `chart_${id}.${format === 'csv' ? 'csv' : 'xlsx'}`;
            downloadBlob(blob, filename);
        });
    },
};

// 下载Blob文件
function downloadBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
}
