import React, { useState, useEffect } from 'react';
import { Modal, Transfer, message } from 'antd';
import { roleAPI, type Role } from '../../api/permission';

interface UserRoleAssignProps {
    visible: boolean;
    userId: string;
    onClose: () => void;
}

const UserRoleAssign: React.FC<UserRoleAssignProps> = ({ visible, userId, onClose }) => {
    const [allRoles, setAllRoles] = useState<Role[]>([]);
    const [userRoleIds, setUserRoleIds] = useState<string[]>([]);
    const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
    const [loading, setLoading] = useState(false);

    // 加载所有角色和用户已有角色
    useEffect(() => {
        if (visible && userId) {
            loadData();
        }
    }, [visible, userId]);

    const loadData = async () => {
        setLoading(true);
        try {
            const [roles, userRoles] = await Promise.all([
                roleAPI.list(),
                roleAPI.getUserRoles(userId),
            ]);

            setAllRoles(roles);
            setUserRoleIds(userRoles.map(r => r.id));
            setSelectedKeys([]);
        } catch (error) {
            message.error('加载数据失败');
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setLoading(true);
        try {
            // 找出需要添加和移除的角色
            const toAdd = userRoleIds.filter(id => !selectedKeys.includes(id));
            const toRemove = selectedKeys.filter(id => !userRoleIds.includes(id));

            // 批量操作
            const promises: Promise<any>[] = [];

            toAdd.forEach(roleId => {
                promises.push(roleAPI.assignToUser(roleId, userId));
            });

            toRemove.forEach(roleId => {
                promises.push(roleAPI.removeFromUser(roleId, userId));
            });

            await Promise.all(promises);
            message.success('保存成功');
            onClose();
        } catch (error) {
            message.error('保存失败');
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal
            title="分配角色"
            open={visible}
            onOk={handleSave}
            onCancel={onClose}
            confirmLoading={loading}
            width={600}
        >
            <Transfer
                dataSource={allRoles.map(role => ({
                    key: role.id,
                    title: role.name,
                    description: role.description,
                }))}
                targetKeys={userRoleIds}
                selectedKeys={selectedKeys}
                onChange={setUserRoleIds}
                onSelectChange={(sourceSelectedKeys, targetSelectedKeys) => {
                    setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys]);
                }}
                render={item => item.title}
                listStyle={{
                    width: 250,
                    height: 400,
                }}
            />
        </Modal>
    );
};

export default UserRoleAssign;
