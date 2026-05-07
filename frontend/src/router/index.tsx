import { Routes, Route } from 'react-router-dom'
import LoginPage from '@/pages/login'

export default function Router() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<div>工作台（建设中）</div>} />
    </Routes>
  )
}
