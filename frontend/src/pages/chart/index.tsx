import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Table,
    Button,
    Space,
    message,
    Popconfirm,
    Card,
    Input,
    Tag,
} from 'antd';
import {
    PlusOutlined,
    EditOutlined,
    DeleteOutlined,
    SearchOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { chartAPI } from '../../api/chart';
import type { ChartView } from '../../types/chart';

const ChartList = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [charts, setCharts] = useState<ChartView[]>([]);
    const [searchText, setSearchText] = useState('');

    // 加载图表列表
    const loadCharts = async () => {
        try {
            setLoading(true);
            const data = await chartAPI.list();
            setCharts(data);
        } catch (error: any) {
            message.error(error.message || '加载图表列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadCharts();
    }, []);

    // 删除图表
    const handleDelete = async (id: string) => {
        try {
            await chartAPI.delete(id);
            message.success('删除成功');
            loadCharts();
        } catch (error: any) {
            message.error(error.message || '删除失败');
        }
    };

    // 图表类型颜色映射
    const getChartTypeColor = (type: string) => {
        const colorMap: Record<string, string> = {
            bar: 'blue',
            line: 'green',
            pie: 'orange',
            scatter: 'purple',
            area: 'cyan',
            table: 'geekblue',
            map: 'red',
            gauge: 'magenta',
            radar: 'lime',
            funnel: 'gold',
            wordcloud: 'volcano',
        };
        return colorMap[type] || 'default';
    };

    // 表格列定义
    const columns: ColumnsType<ChartView> = [
        {
            title: 'ID',
            dataIndex: 'id',
            key: 'id',
            width: 100,
            render: (text: string) => (
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                    {text.slice(0, 8)}...
                </span>
            ),
        },
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            filteredValue: searchText ? [searchText] : null,
            onFilter: (value, record) =>
                record.name.toLowerCase().includes(value.toString().toLowerCase()),
        },
        {
            title: '标题',
            dataIndex: 'title',
            key: 'title',
            render: (text: string) => text || '-',
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            width: 100,
            render: (type: string) => (
                <Tag color={getChartTypeColor(type)}>{type.toUpperCase()}</Tag>
            ),
        },
        {
            title: '数据集ID',
            dataIndex: 'tableId',
            key: 'tableId',
            width: 120,
            render: (text: string) => (
                <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                    {text}
                </span>
            ),
        },
        {
            title: '创建时间',
            dataIndex: 'createTime',
            key: 'createTime',
            width: 180,
            render: (time: number) =>
                time ? new Date(time).toLocaleString('zh-CN') : '-',
            sorter: (a, b) => (a.createTime || 0) - (b.createTime || 0),
        },
        {
            title: '操作',
            key: 'action',
            width: 150,
            fixed: 'right',
            render: (_, record) => (
                <Space>
                    <Button
                        type="link"
                        size="small"
                        icon={<EditOutlined />}
                        onClick={() => navigate(`/chart/edit/${record.id}`)}
                    >
                        编辑
                    </Button>
                    <Popconfirm
                        title="确认删除"
                        description="确定要删除这个图表吗？"
                        onConfirm={() => handleDelete(record.id)}
                        okText="确定"
                        cancelText="取消"
                    >
                        <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                            删除
                        </Button>
                    </Popconfirm>
                </Space>
            ),
        },
    ];

    // 过滤后的数据
    const filteredCharts = searchText
        ? charts.filter((chart) =>
            chart.name.toLowerCase().includes(searchText.toLowerCase())
        )
        : charts;

    return (
        <div style={{ padding: '24px' }}>
            <Card>
                <div
                    style={{
                        marginBottom: 16,
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                    }}
                >
                    <Input
                        placeholder="搜索图表名称"
                        prefix={<SearchOutlined />}
                        value={searchText}
                        onChange={(e) => setSearchText(e.target.value)}
                        style={{ width: 300 }}
                        allowClear
                    />
                    <Button
                        type="primary"
                        icon={<PlusOutlined />}
                        onClick={() => navigate('/chart/create')}
                    >
                        创建图表
                    </Button>
                </div>

                <Table
                    columns={columns}
                    dataSource={filteredCharts}
                    rowKey="id"
                    loading={loading}
                    pagination={{
                        pageSize: 10,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total) => `共 ${total} 条`,
                    }}
                    scroll={{ x: 1200 }}
                />
            </Card>
        </div>
    );
};

export default ChartList;
