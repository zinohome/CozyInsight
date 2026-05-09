import { useParams } from 'react-router-dom'

export default function ScreenView() {
  const { id } = useParams<{ id: string }>()
  return <div>数据大屏预览 #{id}</div>
}
