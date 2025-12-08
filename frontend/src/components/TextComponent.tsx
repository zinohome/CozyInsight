import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

export interface TextComponentProps {
    content?: string;
    editable?: boolean;
    onChange?: (content: string) => void;
}

/**
 * 文本组件 - 用于仪表板中显示文本内容
 */
export const TextComponent: React.FC<TextComponentProps> = ({
    content = '双击编辑文本',
    editable = false,
    onChange,
}) => {
    const [editing, setEditing] = React.useState(false);
    const [text, setText] = React.useState(content);

    const handleDoubleClick = () => {
        if (editable) {
            setEditing(true);
        }
    };

    const handleChange = (newText: string) => {
        setText(newText);
        if (onChange) {
            onChange(newText);
        }
    };

    const handleBlur = () => {
        setEditing(false);
    };

    return (
        <div
            style={{ padding: 16, height: '100%', overflow: 'auto' }}
            onDoubleClick={handleDoubleClick}
        >
            {editing ? (
                <textarea
                    value={text}
                    onChange={(e) => handleChange(e.target.value)}
                    onBlur={handleBlur}
                    autoFocus
                    style={{
                        width: '100%',
                        height: '100%',
                        fontSize: 14,
                        border: '1px solid #d9d9d9',
                        padding: 8,
                        borderRadius: 4,
                    }}
                />
            ) : (
                <div dangerouslySetInnerHTML={{ __html: text }} />
            )}
        </div>
    );
};

export default TextComponent;
