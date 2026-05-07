import React, { useState } from 'react';
import { Modal, Form, Input, Switch, DatePicker, message, Select } from 'antd';
import { shareAPI } from '../../api/share';

interface ShareDialogProps {
    visible: boolean;
    resourceType: 'dashboard' | 'chart';
    resourceId: string;
    onClose: () => void;
}

const ShareDialog: React.FC<ShareDialogProps> = ({ visible, resourceType, resourceId, onClose }) => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);
    const [shareLink, setShareLink] = useState<string>('');

    const handleCreate = async () => {
        try {
            const values = await form.validateFields();
            setLoading(true);

            const expireTime = values.expireTime
                ? values.expireTime.valueOf()
                : Date.now() + 7 * 24 * 60 * 60 * 1000; // 默认7天

            const share = await shareAPI.create({
                resourceType,
                resourceId,
                password: values.enablePassword ? values.password : '',
                expireTime,
            });

            const link = `${window.location.origin}/share/${share.token}`;
            setShareLink(link);
            message.success('分享链接已生成');
        } catch (error) {
            message.error('创建分享失败');
        } finally {
            setLoading(false);
        }
    };

    const handleCopy = () => {
        navigator.clipboard.writeText(shareLink);
        message.success('已复制到剪贴板');
    };

    return (
        <Modal
            title="生成分享链接"
            open={visible}
            onOk={handleCreate}
            onCancel={onClose}
            confirmLoading={loading}
            okText={shareLink ? '关闭' : '生成链接'}
            onOk={shareLink ? onClose : handleCreate}
        >
            {!shareLink ? (
                <Form form={form} layout="vertical">
                    <Form.Item
                        label="有效期"
                        name="expireTime"
                    >
                        <DatePicker showTime placeholder="选择过期时间(默认7天)" style={{ width: '100%' }} />
                    </Form.Item>

                    <Form.Item
                        label="密码保护"
                        name="enablePassword"
                        valuePropName="checked"
                    >
                        <Switch />
                    </Form.Item>

                    <Form.Item
                        noStyle
                        shouldUpdate={(prev, curr) => prev.enablePassword !== curr.enablePassword}
                    >
                        {({ getFieldValue }) =>
                            getFieldValue('enablePassword') ? (
                                <Form.Item
                                    label="访问密码"
                                    name="password"
                                    rules={[{ required: true, message: '请输入密码' }]}
                                >
                                    <Input.Password placeholder="请输入密码" />
                                </Form.Item>
                            ) : null
                        }
                    </Form.Item>
                </Form>
            ) : (
                <div>
                    <p>分享链接:</p>
                    <Input.Search
                        value={shareLink}
                        readOnly
                        enterButton="复制"
                        onSearch={handleCopy}
                    />
                </div>
            )}
        </Modal>
    );
};

export default ShareDialog;
