import React from 'react';
import { Map } from '@ant-design/maps';

interface MapChartProps {
    data: any[];
    config: {
        longitudeField: string;
        latitudeField: string;
        colorField?: string;
        sizeField?: string;
        title?: string;
    };
}

const MapChart: React.FC<MapChartProps> = ({ data, config }) => {
    // 转换数据格式
    const mapData = data.map(item => ({
        lng: Number(item[config.longitudeField]),
        lat: Number(item[config.latitudeField]),
        value: config.sizeField ? Number(item[config.sizeField]) : 1,
        name: config.colorField ? item[config.colorField] : '',
    }));

    const mapConfig = {
        map: {
            type: 'mapbox',
            style: 'light',
            center: [120.19, 30.26],
            zoom: 4,
        },
        source: {
            data: mapData,
            parser: {
                type: 'json',
                x: 'lng',
                y: 'lat',
            },
        },
        size: config.sizeField ? {
            field: 'value',
            value: [4, 40],
        } : 4,
        color: config.colorField ? {
            field: 'name',
            value: ['#5B8FF9', '#5AD8A6', '#5D7092', '#F6BD16'],
        } : '#5B8FF9',
        style: {
            opacity: 0.8,
            strokeWidth: 1,
        },
        tooltip: {
            showTitle: true,
            items: [
                {
                    field: 'name',
                    alias: '名称',
                },
                {
                    field: 'value',
                    alias: '数值',
                },
            ],
        },
    };

    return (
        <div style={{ height: '100%', minHeight: 400 }}>
            {/* 注意: @ant-design/maps 需要单独安装和配置 */}
            {/* 这里提供基础结构,实际使用需要配置地图token */}
            <div style={{ textAlign: 'center', padding: '100px 0' }}>
                <p>地图图表 (需要配置地图服务)</p>
                <p>经度字段: {config.longitudeField}</p>
                <p>纬度字段: {config.latitudeField}</p>
                <p>数据点数: {data.length}</p>
            </div>
        </div>
    );
};

export default MapChart;
