import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import Router from './router'

dayjs.locale('zh-cn')

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Router />
    </ConfigProvider>
  )
}

export default App
