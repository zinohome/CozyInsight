import React, { useEffect, useState } from 'react';
import { Layout, Tree, Empty, Button, Card, Space } from 'antd';
import { PlusOutlined, FolderAddOutlined, TableOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import { getDatasetGroups } from '../../api/dataset';
import type { DatasetGroup } from '../../types/dataset';

const { Sider, Content } = Layout;

const DatasetList: React.FC = () => {
    const [treeData, setTreeData] = useState<DataNode[]>([]);
    const [loading, setLoading] = useState(false);
    const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

    const fetchGroups = async () => {
        try {
            setLoading(true);
            const groups = await getDatasetGroups();
            // Transform groups to tree data
            // This is a simplified transformation. In real app, we need to build the tree from flat list
            const buildTree = (items: DatasetGroup[], pid: string | null = null): DataNode[] => {
                return items
                    .filter(item => item.pid === (pid || '')) // Assuming empty string for root pid
                    .map(item => ({
                        key: item.id,
                        title: item.name,
                        icon: item.nodeType === 'folder' ? <FolderAddOutlined /> : <TableOutlined />,
                        children: buildTree(items, item.id),
                        isLeaf: item.nodeType === 'dataset',
                    }));
            };
            // For now, just show flat list as root nodes if pid handling is complex
            // Or just map all for demo
            const nodes = groups.map(g => ({
                key: g.id,
                title: g.name,
                icon: g.nodeType === 'folder' ? <FolderAddOutlined /> : <TableOutlined />,
                isLeaf: g.nodeType === 'dataset',
            }));
            setTreeData(nodes);
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchGroups();
    }, []);

    return (
        <Layout style={{ background: '#fff', padding: 24, minHeight: 360 }}>
            <Sider width={280} style={{ background: '#fff', borderRight: '1px solid #f0f0f0' }}>
                <div style={{ padding: '0 16px 16px', borderBottom: '1px solid #f0f0f0', marginBottom: 16 }}>
                    <Space>
                        <Button type="primary" icon={<PlusOutlined />} size="small" loading={loading}>新建数据集</Button>
                        <Button icon={<FolderAddOutlined />} size="small" disabled={loading}>新建分组</Button>
                    </Space>
                </div>
                <Tree
                    treeData={treeData}
                    onSelect={(keys) => setSelectedKeys(keys)}
                    blockNode
                    disabled={loading}
                />
            </Sider>
            <Content style={{ padding: '0 24px' }}>
                {selectedKeys.length > 0 ? (
                    <Card title="数据集详情">
                        <p>Selected ID: {selectedKeys[0]}</p>
                        {/* Table list or details would go here */}
                    </Card>
                ) : (
                    <Empty description="请选择左侧数据集或分组" style={{ marginTop: 100 }} />
                )}
            </Content>
        </Layout>
    );
};

export default DatasetList;
