import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'
// ✅ Импортируем qualityLevels ГЛОБАЛЬНО (один раз)
import 'videojs-contrib-quality-levels';
ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
