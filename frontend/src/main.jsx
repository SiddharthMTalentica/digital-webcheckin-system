import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './index.css'
import Homepage from './components/Homepage'
import App from './App.jsx'
import WebCheckInApp from './components/webcheckin/WebCheckInApp'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Homepage />} />
        <Route path="/flight-booking-system" element={<App />} />
        <Route path="/web-check-in" element={<WebCheckInApp />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
