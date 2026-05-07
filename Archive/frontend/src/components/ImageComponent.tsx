import React, { useState } from 'react';
import { Upload, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { UploadFile, UploadProps } from 'antd';

export interface ImageComponentProps {
    url?: string;
    editable?: boolean;
    onChange?: (url: string) => void;
}

/**
 * 图片组件 - 用于仪表板中显示图片
 */
export const ImageComponent: React.FC<ImageComponentProps> = ({
    url,
    editable = false,
    onChange,
}) => {
    const [imageUrl, setImageUrl] = useState(url);

    const handleUpload: UploadProps['customRequest'] = async (options) => {
        const { file, onSuccess, onError } = options;

        try {
            // TODO: 实际上传到服务器
            // const formData = new FormData();
            // formData.append('file', file);
            // const response = await fetch('/api/upload', {
            //   method: 'POST',
            //   body: formData,
            // });
            // const data = await response.json();

            // 模拟上传成功
            const reader = new FileReader();
            reader.onload = (e) => {
                const url = e.target?.result as string;
                setImageUrl(url);
                if (onChange) {
                    onChange(url);
                }
                message.success('图片上传成功');
                if (onSuccess) {
                    onSuccess('ok');
                }
            };
            reader.readAsDataURL(file as File);
        } catch (error) {
            message.error('图片上传失败');
            if (onError) {
                onError(error as Error);
            }
        }
    };

    const uploadButton = (
        <div>
            <PlusOutlined />
            <div style={{ marginTop: 8 }}>上传图片</div>
        </div>
    );

    if (!editable && imageUrl) {
        return (
            <div
                style={{
                    width: '100%',
                    height: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    overflow: 'hidden',
                }}
            >
                <img
                    src={imageUrl}
                    alt="Dashboard"
                    style={{
                        maxWidth: '100%',
                        maxHeight: '100%',
                        objectFit: 'contain',
                    }}
                />
            </div>
        );
    }

    return (
        <div
            style={{
                width: '100%',
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
            }}
        >
            {editable ? (
                <Upload
                    name="image"
                    listType="picture-card"
                    showUploadList={false}
                    customRequest={handleUpload}
                    beforeUpload={(file) => {
                        const isImage = file.type.startsWith('image/');
                        if (!isImage) {
                            message.error('只能上传图片文件！');
                        }
                        const isLt2M = file.size / 1024 / 1024 < 2;
                        if (!isLt2M) {
                            message.error('图片大小不能超过 2MB！');
                        }
                        return isImage && isLt2M;
                    }}
                >
                    {imageUrl ? (
                        <img src={imageUrl} alt="Dashboard" style={{ width: '100%' }} />
                    ) : (
                        uploadButton
                    )}
                </Upload>
            ) : (
                <div style={{ color: '#999' }}>未设置图片</div>
            )}
        </div>
    );
};

export default ImageComponent;
