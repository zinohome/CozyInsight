import { Button, Empty, Skeleton, Spin } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'

interface ChartEmptyProps {
  height?: number
  message?: string
}

export function ChartEmpty({ height = 300, message = '暂无数据' }: ChartEmptyProps) {
  return (
    <div
      style={{
        height,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#fafafa',
        border: '1px dashed #e0e0e0',
        borderRadius: 4,
      }}
    >
      <Empty description={message} />
    </div>
  )
}

interface ChartLoadingProps {
  height?: number
}

export function ChartLoading({ height = 300 }: ChartLoadingProps) {
  return (
    <div
      style={{
        height,
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
        padding: 16,
        background: '#fafafa',
        borderRadius: 4,
      }}
    >
      <Skeleton active paragraph={{ rows: 2 }} />
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', flex: 1 }}>
        <Spin tip="加载中…" />
      </div>
    </div>
  )
}

interface ChartErrorProps {
  height?: number
  error: string
  onRetry?: () => void
}

export function ChartError({ height = 300, error, onRetry }: ChartErrorProps) {
  return (
    <div
      style={{
        height,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 12,
        background: '#fff2f0',
        border: '1px solid #ffccc7',
        borderRadius: 4,
        padding: 16,
      }}
    >
      <div style={{ color: '#cf1322', fontWeight: 500 }}>加载失败</div>
      <div
        style={{
          color: '#888',
          fontSize: 12,
          maxWidth: '80%',
          textAlign: 'center',
          wordBreak: 'break-all',
        }}
      >
        {error}
      </div>
      {onRetry && (
        <Button icon={<ReloadOutlined />} size="small" onClick={onRetry}>
          重试
        </Button>
      )}
    </div>
  )
}
