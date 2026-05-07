import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'
import DatasourcePage from '@/pages/datasource'
import DatasetPage from '@/pages/dataset'
import ChartPage from '@/pages/chart'
import DashboardPage from '@/pages/dashboard'
import UserPage from '@/pages/system/user'
import RolePage from '@/pages/system/role'
import LogPage from '@/pages/system/log'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div style={{ padding: 24 }}>工作台（建设中）</div>} />
      <Route path="/datasource" element={<DatasourcePage />} />
      <Route path="/dataset" element={<DatasetPage />} />
      <Route path="/chart" element={<ChartPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
      <Route path="/system/user" element={<UserPage />} />
      <Route path="/system/role" element={<RolePage />} />
      <Route path="/system/log" element={<LogPage />} />
    </Routes>
  )
}
