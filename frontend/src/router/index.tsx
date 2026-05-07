import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'
import DatasourcePage from '@/pages/datasource'
import DatasetPage from '@/pages/dataset'
import ChartPage from '@/pages/chart'
import DashboardPage from '@/pages/dashboard'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>工作台（建设中）</div>} />
      <Route path="/datasource" element={<DatasourcePage />} />
      <Route path="/dataset" element={<DatasetPage />} />
      <Route path="/chart" element={<ChartPage />} />
      <Route path="/dashboard" element={<DashboardPage />} />
    </Routes>
  )
}
