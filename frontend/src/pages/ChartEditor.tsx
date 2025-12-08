import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Space, message } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import ChartRenderer from '../components/ChartRenderer';
import ChartConfigPanel from '../components/ChartConfigPanel';
import ChartFilterPanel, { ChartFilter } from '../components/ChartFilterPanel';
import DataPreviewTable from '../components/DataPreviewTable';
import { chartApi, datasetApi } from '../services/api';
import type { ChartConfig, DatasetField } from '../types/chart';

/**
 * 图表编辑器页面
 */
const ChartEditor: React.FC = () => {
    return;
}

if (!chartConfig.xField || !chartConfig.yField) {
    message.warning('请配置图表字段');
    return;
}

setSaving(true);
try {
    const chartData = {
        name: chartName,
        tableId: datasetId,
        type: chartConfig.type || 'column',
        xAxis: JSON.stringify({
            fields: [{ name: chartConfig.xField }],
        }),
        yAxis: JSON.stringify({
            fields: [{ name: chartConfig.yField }],
        }),
    };

    if (id) {
        await chartApi.update(id, chartData);
        message.success('保存成功');
    } else {
        const newChart = await chartApi.create(chartData);
        message.success('创建成功');
        navigate(`/chart/edit/${newChart.id}`);
    }
} catch (error) {
    message.error('保存失败');
    console.error(error);
} finally {
    setSaving(false);
}
    };

const renderChartPreview = () => {
    if (!chartConfig.xField || !chartConfig.yField || chartData.length === 0) {
        return (
            <div
                style={{
                    height: 400,
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    background: '#fafafa',
                    border: '1px dashed #d9d9d9',
                    borderRadius: 4,
                }}
            >
                <div style={{ textAlign: 'center', color: '#999' }}>
                    <EyeOutlined style={{ fontSize: 48, marginBottom: 16 }} />
                    <div>配置图表字段后可预览</div>
                </div>
            </div>
        );
    }

    const config = {
        ...chartConfig,
        data: chartData,
    };

    return <ChartRenderer config={config as any} style={{ height: 400 }} />;
};

if (loading) {
    return (
        <div style={{ textAlign: 'center', padding: '100px 0' }}>
            <Spin size="large" tip="加载中..." />
        </div>
    );
}

return (
    <div style={{ padding: 24 }}>
        <Card
            title={chartName}
            extra={
                <Space>
                    <Button onClick={() => navigate('/chart')}>取消</Button>
                    <Button
                        type="primary"
                        icon={<SaveOutlined />}
                        loading={saving}
                        onClick={handleSave}
                    >
                        保存
                    </Button>
                </Space>
            }
        >
            <Row gutter={16}>
                {/* 左侧：配置面板 */}
                <Col span={6}>
                    <ChartConfigPanel
                        fields={fields}
                        onConfigChange={handleConfigChange}
                        initialConfig={chartConfig}
                    />
                </Col>

                {/* 中间：图表预览 */}
                <Col span={18}>
                    <Card title="图表预览" size="small" style={{ marginBottom: 16 }}>
                        {renderChartPreview()}
                    </Card>

                    {/* 数据预览 */}
                    {datasetId && <DataPreviewTable datasetId={datasetId} limit={50} />}
                </Col>
            </Row>
        </Card>
    </div>
);
};

export default ChartEditor;
